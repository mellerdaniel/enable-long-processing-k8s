// Package admission handles kubernetes admissions,
// it takes admission requests and returns admission reviews;
// for example, to mutate or validate pods
package admission

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/slackhq/simple-kubernetes-webhook/pkg/checkpoint"
	"github.com/slackhq/simple-kubernetes-webhook/pkg/mutation"
	"github.com/slackhq/simple-kubernetes-webhook/pkg/validation"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"net/http"
)

// Admitter is a container for admission business
type Admitter struct {
	Logger  *logrus.Entry
	Request *admissionv1.AdmissionRequest
}

// MutatePodReview takes an admission request and mutates the pod within,
// it returns an admission review with mutations as a json patch (if any)
func (a Admitter) MutatePodReview() (*admissionv1.AdmissionReview, error) {
	pod, err := a.Pod()
	podName := a.getPodName()
	m := mutation.NewMutator(a.Logger)
	checkpoint.GetCheckpointManager(a.Logger).PodCreated(podName)
	patch, err := m.MutatePodPatch(pod)
	if err != nil {
		e := fmt.Sprintf("could not mutate pod: %v", err)
		return reviewResponse(a.Request.UID, false, http.StatusBadRequest, e), err
	}

	return patchReviewResponse(a.Request.UID, patch)
}

func (a Admitter) DeletePodReview() (*admissionv1.AdmissionReview, error) {
	podName := a.getPodName()
	checkpoint.GetCheckpointManager(a.Logger).PodDeleted(podName)
	return &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     a.Request.UID,
			Allowed: true,
		},
	}, nil
}

// MutatePodReview takes an admission request and validates the pod within
// it returns an admission review
func (a Admitter) ValidatePodReview() (*admissionv1.AdmissionReview, error) {
	pod, err := a.Pod()
	if err != nil {
		e := fmt.Sprintf("could not parse pod in admission review request: %v", err)
		return reviewResponse(a.Request.UID, false, http.StatusBadRequest, e), err
	}

	v := validation.NewValidator(a.Logger)
	val, err := v.ValidatePod(pod)
	if err != nil {
		e := fmt.Sprintf("could not validate pod: %v", err)
		return reviewResponse(a.Request.UID, false, http.StatusBadRequest, e), err
	}

	if !val.Valid {
		return reviewResponse(a.Request.UID, false, http.StatusForbidden, val.Reason), nil
	}

	return reviewResponse(a.Request.UID, true, http.StatusAccepted, "valid pod"), nil
}

// Pod extracts a pod from an admission request
func (a Admitter) Pod() (*corev1.Pod, error) {
	if a.Request.Kind.Kind != "Pod" {
		return nil, fmt.Errorf("only pods are supported here")
	}
	op := a.Request.Operation
	p := corev1.Pod{}
	switch op {
	case admissionv1.Create:
		if err := json.Unmarshal(a.Request.Object.Raw, &p); err != nil {
			return nil, err
		}
	case admissionv1.Delete:
		if err := json.Unmarshal(a.Request.OldObject.Raw, &p); err != nil {
			return nil, err
		}

	}
	return &p, nil
}

// reviewResponse TODO: godoc
func reviewResponse(uid types.UID, allowed bool, httpCode int32,
	reason string) *admissionv1.AdmissionReview {
	return &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     uid,
			Allowed: allowed,
			Result: &metav1.Status{
				Code:    httpCode,
				Message: reason,
			},
		},
	}
}

// patchReviewResponse builds an admission review with given json patch
func patchReviewResponse(uid types.UID, patch []byte) (*admissionv1.AdmissionReview, error) {
	patchType := admissionv1.PatchTypeJSONPatch

	return &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:       uid,
			Allowed:   true,
			PatchType: &patchType,
			Patch:     patch,
		},
	}, nil
}

func (a Admitter) getPodName() string {
	pod, err := a.Pod()
	if err != nil {
		e := fmt.Sprintf("error in getPodName, could not parse pod in admission review request: %v", err)
		a.Logger.Error(e)
		return ""
	}
	return fmt.Sprintf("%v_%v_%v", a.Request.Namespace, a.Request.Name, pod.Spec.Containers[0].Name)
}
