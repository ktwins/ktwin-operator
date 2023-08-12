package eventStore

import (
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev0 "ktwin/operator/api/core/v0"
	dtdv0 "ktwin/operator/api/dtd/v0"

	kserving "knative.dev/serving/pkg/apis/serving/v1"
)

const (
	EVENT_STORE_SERVICE string = "event-store"
)

func NewEventStore() EventStore {
	return &eventStore{}
}

type EventStore interface {
	GetEventStoreService(eventStore *corev0.EventStore) *kserving.Service
	CreateTwinInterface(twinInterface *dtdv0.TwinInstance) error
	DeleteTwinInterface(twinInterface *dtdv0.TwinInstance) error
	CreateTwinInstance(twinInstance *dtdv0.TwinInstance) error
	DeleteTwinInstance(twinInstance *dtdv0.TwinInstance) error
}

type eventStore struct{}

func (t *eventStore) GetEventStoreService(eventStore *corev0.EventStore) *kserving.Service {
	eventStoreName := eventStore.ObjectMeta.Name

	service := &kserving.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      eventStore.ObjectMeta.Name,
			Namespace: eventStore.ObjectMeta.Namespace,
			Labels: map[string]string{
				"ktwin/event-store": eventStoreName,
			},
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: eventStore.APIVersion,
					Kind:       eventStore.Kind,
					Name:       eventStore.ObjectMeta.Name,
					UID:        eventStore.UID,
				},
			},
		},
		Spec: kserving.ServiceSpec{
			ConfigurationSpec: kserving.ConfigurationSpec{
				Template: kserving.RevisionTemplateSpec{
					ObjectMeta: v1.ObjectMeta{
						Name: eventStoreName + "-v1",
						Annotations: map[string]string{
							"autoscaling.knative.dev/target":   "2",
							"autoscaling.knative.dev/minScale": "1",
							"autoscaling.knative.dev/maxScale": "10",
						},
					},
					Spec: kserving.RevisionSpec{
						PodSpec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:            "ktwin-event-store",
									Image:           "dev.local/ktwin/" + "event-store" + ":0.1",
									ImagePullPolicy: corev1.PullIfNotPresent,
									Env: []corev1.EnvVar{
										{
											Name:  "DB_HOST",
											Value: "scylla-client.scylla.svc.cluster.local",
										},
										{
											Name:  "DB_PASSWORD",
											Value: "",
										},
										{
											Name:  "DB_KEYSPACE",
											Value: "ktwin",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return service
}

func (t *eventStore) CreateTwinInterface(twinInterface *dtdv0.TwinInstance) error {
	// TwinInstance

	// Interface
	return nil
}

func (t *eventStore) DeleteTwinInterface(twinInterface *dtdv0.TwinInstance) error {
	// TwinInstance

	return nil
}

func (t *eventStore) CreateTwinInstance(twinInstance *dtdv0.TwinInstance) error {
	//

	return nil
}

func (t *eventStore) DeleteTwinInstance(twinInstance *dtdv0.TwinInstance) error {
	//

	return nil
}
