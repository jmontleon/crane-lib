package state_transfer

import (
	"context"
	"strconv"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const ()

func (r *RsyncTransfer) CreateClient(c client.Client) error {
	err := createRsyncClientResources(c, r)
	if err != nil {
		return err
	}

	transport, err := CreateTransportClient(r.Transport(), c, r)
	if err != nil {
		return err
	}
	r.SetTransport(transport)

	err = createRsyncClient(c, r)
	if err != nil {
		return err
	}

	return nil
}

func createRsyncClientResources(c client.Client, r *RsyncTransfer) error {
	// no resource are created for rsync client side
	return nil
}

func createRsyncClient(c client.Client, r *RsyncTransfer) error {
	podLabels := labels
	podLabels["pvc"] = r.PVC().Name
	containers := []v1.Container{
		{
			Name:  "rsync",
			Image: rsyncImage,
			Command: []string{
				"/usr/bin/rsync",
				"-vvvv",
				"--delete",
				"--recursive",
				"--compress",
				"rsync://" + r.Username() + "@" + connectionHostname(r) + ":" + strconv.Itoa(int(connectionPort(r))) + "/mnt",
				"/mnt/",
			},
			Env: []v1.EnvVar{
				{
					Name:  "RSYNC_PASSWORD",
					Value: r.Password(),
				},
			},

			VolumeMounts: []v1.VolumeMount{
				{
					Name:      "mnt",
					MountPath: "/mnt",
				},
			},
		},
	}

	for _, container := range r.Transport().ClientContainers() {
		containers = append(containers, container)
	}

	volumes := []v1.Volume{
		{
			Name: "mnt",
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: r.PVC().Name,
				},
			},
		},
	}

	for _, volume := range r.Transport().ClientVolumes() {
		volumes = append(volumes, volume)
	}

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.PVC().Name,
			Namespace: r.PVC().Namespace,
			Labels:    podLabels,
		},
		Spec: v1.PodSpec{
			Containers:    containers,
			Volumes:       volumes,
			RestartPolicy: v1.RestartPolicyOnFailure,
		},
	}

	return c.Create(context.TODO(), &pod, &client.CreateOptions{})
}
