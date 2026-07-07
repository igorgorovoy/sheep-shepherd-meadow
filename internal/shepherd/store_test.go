package shepherd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeleteDeploymentCascadesPods(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(filepath.Join(dir, "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })

	dep := &Deployment{
		Kind: "Deployment",
		Metadata: ObjectMeta{
			Name:      "wordpress",
			Namespace: "default",
		},
		Spec: DeploymentSpec{
			Replicas: 2,
			Selector: map[string]string{"app": "wordpress"},
			Template: PodTemplate{
				Metadata: ObjectMeta{
					Labels: map[string]string{"app": "wordpress"},
				},
			},
		},
	}
	if err := store.CreateDeployment(dep); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"wordpress-0", "wordpress-1"} {
		pod := &Pod{
			Kind: "Pod",
			Metadata: ObjectMeta{
				Name:      name,
				Namespace: "default",
				Labels:    map[string]string{"app": "wordpress"},
			},
		}
		if err := store.CreatePod(pod); err != nil {
			t.Fatal(err)
		}
	}

	// Unrelated pod must survive cascade delete.
	other := &Pod{
		Kind: "Pod",
		Metadata: ObjectMeta{
			Name:      "manual-pod",
			Namespace: "default",
			Labels:    map[string]string{"app": "manual"},
		},
	}
	if err := store.CreatePod(other); err != nil {
		t.Fatal(err)
	}

	if err := store.DeleteDeployment("default", "wordpress"); err != nil {
		t.Fatal(err)
	}

	if _, err := store.GetDeployment("default", "wordpress"); err == nil {
		t.Fatal("expected deployment to be deleted")
	}

	pods, err := store.ListPods("default")
	if err != nil {
		t.Fatal(err)
	}
	if len(pods) != 1 {
		t.Fatalf("expected 1 pod left, got %d", len(pods))
	}
	if pods[0].Metadata.Name != "manual-pod" {
		t.Fatalf("unexpected remaining pod: %s", pods[0].Metadata.Name)
	}
}

func TestDeleteDeploymentNotFound(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(filepath.Join(dir, "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close(); os.RemoveAll(dir) })

	if err := store.DeleteDeployment("default", "missing"); err == nil {
		t.Fatal("expected error deleting missing deployment")
	}
}
