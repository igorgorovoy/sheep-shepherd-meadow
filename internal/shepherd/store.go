package shepherd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketPods        = []byte("pods")
	bucketServices    = []byte("services")
	bucketDeployments = []byte("deployments")
	bucketNodes       = []byte("nodes")
	bucketEvents      = []byte("events")
)

type Store struct {
	db *bolt.DB
	mu sync.RWMutex

	// Watch channels for controllers
	podWatchers        []chan Event
	deploymentWatchers []chan Event
	watchMu            sync.Mutex
}

func NewStore(path string) (*Store, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}

	// Create buckets
	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{bucketPods, bucketServices, bucketDeployments, bucketNodes, bucketEvents} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// --- Pod Operations ---

func (s *Store) CreatePod(pod *Pod) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := nsKey(pod.Metadata.Namespace, pod.Metadata.Name)
	return s.put(bucketPods, key, pod)
}

func (s *Store) UpdatePod(pod *Pod) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := nsKey(pod.Metadata.Namespace, pod.Metadata.Name)
	if err := s.put(bucketPods, key, pod); err != nil {
		return err
	}

	s.notify(s.podWatchers, Event{
		Type:      "Normal",
		Reason:    "Updated",
		Object:    "pod/" + pod.Metadata.Name,
		Timestamp: time.Now(),
	})

	return nil
}

func (s *Store) GetPod(namespace, name string) (*Pod, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var pod Pod
	if err := s.get(bucketPods, nsKey(namespace, name), &pod); err != nil {
		return nil, err
	}
	return &pod, nil
}

func (s *Store) ListPods(namespace string) ([]*Pod, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var pods []*Pod
	err := s.list(bucketPods, namespace, func(data []byte) error {
		var pod Pod
		if err := json.Unmarshal(data, &pod); err != nil {
			return err
		}
		pods = append(pods, &pod)
		return nil
	})
	return pods, err
}

func (s *Store) DeletePod(namespace, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.delete(bucketPods, nsKey(namespace, name))
}

// --- Service Operations ---

func (s *Store) CreateService(svc *Service) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := nsKey(svc.Metadata.Namespace, svc.Metadata.Name)
	return s.put(bucketServices, key, svc)
}

func (s *Store) GetService(namespace, name string) (*Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var svc Service
	if err := s.get(bucketServices, nsKey(namespace, name), &svc); err != nil {
		return nil, err
	}
	return &svc, nil
}

func (s *Store) ListServices(namespace string) ([]*Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var svcs []*Service
	err := s.list(bucketServices, namespace, func(data []byte) error {
		var svc Service
		if err := json.Unmarshal(data, &svc); err != nil {
			return err
		}
		svcs = append(svcs, &svc)
		return nil
	})
	return svcs, err
}

func (s *Store) UpdateService(svc *Service) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := nsKey(svc.Metadata.Namespace, svc.Metadata.Name)
	return s.put(bucketServices, key, svc)
}

func (s *Store) DeleteService(namespace, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.delete(bucketServices, nsKey(namespace, name))
}

// --- Deployment Operations ---

func (s *Store) CreateDeployment(dep *Deployment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := nsKey(dep.Metadata.Namespace, dep.Metadata.Name)
	if err := s.put(bucketDeployments, key, dep); err != nil {
		return err
	}

	s.notify(s.deploymentWatchers, Event{
		Type:      "Normal",
		Reason:    "Created",
		Object:    "deployment/" + dep.Metadata.Name,
		Timestamp: time.Now(),
	})

	return nil
}

func (s *Store) GetDeployment(namespace, name string) (*Deployment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var dep Deployment
	if err := s.get(bucketDeployments, nsKey(namespace, name), &dep); err != nil {
		return nil, err
	}
	return &dep, nil
}

func (s *Store) ListDeployments(namespace string) ([]*Deployment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var deps []*Deployment
	err := s.list(bucketDeployments, namespace, func(data []byte) error {
		var dep Deployment
		if err := json.Unmarshal(data, &dep); err != nil {
			return err
		}
		deps = append(deps, &dep)
		return nil
	})
	return deps, err
}

func (s *Store) UpdateDeployment(dep *Deployment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := nsKey(dep.Metadata.Namespace, dep.Metadata.Name)
	return s.put(bucketDeployments, key, dep)
}

func (s *Store) DeleteDeployment(namespace, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := nsKey(namespace, name)
	var dep Deployment
	if err := s.get(bucketDeployments, key, &dep); err != nil {
		return err
	}

	var podNames []string
	if err := s.list(bucketPods, namespace, func(data []byte) error {
		var pod Pod
		if err := json.Unmarshal(data, &pod); err != nil {
			return err
		}
		if matchLabels(pod.Metadata.Labels, dep.Spec.Selector) {
			podNames = append(podNames, pod.Metadata.Name)
		}
		return nil
	}); err != nil {
		return err
	}

	for _, podName := range podNames {
		if err := s.delete(bucketPods, nsKey(namespace, podName)); err != nil {
			return err
		}
	}

	return s.delete(bucketDeployments, key)
}

// --- Node Operations ---

func (s *Store) RegisterNode(node *Node) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.put(bucketNodes, []byte(node.Metadata.Name), node)
}

func (s *Store) GetNode(name string) (*Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var node Node
	if err := s.get(bucketNodes, []byte(name), &node); err != nil {
		return nil, err
	}
	return &node, nil
}

func (s *Store) ListNodes() ([]*Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var nodes []*Node
	err := s.list(bucketNodes, "", func(data []byte) error {
		var node Node
		if err := json.Unmarshal(data, &node); err != nil {
			return err
		}
		nodes = append(nodes, &node)
		return nil
	})
	return nodes, err
}

func (s *Store) UpdateNode(node *Node) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.put(bucketNodes, []byte(node.Metadata.Name), node)
}

func (s *Store) DeleteNode(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.delete(bucketNodes, []byte(name))
}

// --- Watch ---

func (s *Store) WatchPods() chan Event {
	s.watchMu.Lock()
	defer s.watchMu.Unlock()

	ch := make(chan Event, 64)
	s.podWatchers = append(s.podWatchers, ch)
	return ch
}

func (s *Store) WatchDeployments() chan Event {
	s.watchMu.Lock()
	defer s.watchMu.Unlock()

	ch := make(chan Event, 64)
	s.deploymentWatchers = append(s.deploymentWatchers, ch)
	return ch
}

// --- Events ---

func (s *Store) RecordEvent(evt Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%d-%s", evt.Timestamp.UnixNano(), evt.Object)
	return s.put(bucketEvents, []byte(key), evt)
}

func (s *Store) ListEvents(limit int) ([]Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []Event
	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketEvents)
		c := b.Cursor()
		count := 0
		// Iterate from last to first for most recent
		for k, v := c.Last(); k != nil && count < limit; k, v = c.Prev() {
			var evt Event
			if err := json.Unmarshal(v, &evt); err == nil {
				events = append(events, evt)
			}
			count++
		}
		return nil
	})
	return events, nil
}

// --- Internal helpers ---

func (s *Store) put(bucket []byte, key []byte, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucket).Put(key, data)
	})
}

func (s *Store) get(bucket []byte, key []byte, v any) error {
	return s.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(bucket).Get(key)
		if data == nil {
			return fmt.Errorf("not found")
		}
		return json.Unmarshal(data, v)
	})
}

func (s *Store) delete(bucket []byte, key []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucket).Delete(key)
	})
}

func (s *Store) list(bucket []byte, prefix string, fn func([]byte) error) error {
	return s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		return b.ForEach(func(k, v []byte) error {
			if prefix == "" || strings.HasPrefix(string(k), prefix+"/") {
				return fn(v)
			}
			return nil
		})
	})
}

func (s *Store) notify(watchers []chan Event, evt Event) {
	s.watchMu.Lock()
	defer s.watchMu.Unlock()

	for _, ch := range watchers {
		select {
		case ch <- evt:
		default:
		}
	}
}

func nsKey(namespace, name string) []byte {
	if namespace == "" {
		namespace = "default"
	}
	return []byte(namespace + "/" + name)
}

// namespaceFromKey extracts the namespace segment from a store key "ns/name".
func namespaceFromKey(key string) string {
	if i := strings.Index(key, "/"); i >= 0 {
		return key[:i]
	}
	return "default"
}

// ListNamespaces returns sorted unique namespace names found in namespaced
// resource buckets (pods, services, deployments). Always includes "default".
func (s *Store) ListNamespaces() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	seen := map[string]bool{"default": true}
	for _, bucket := range [][]byte{bucketPods, bucketServices, bucketDeployments} {
		if err := s.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(bucket)
			if b == nil {
				return nil
			}
			return b.ForEach(func(k, _ []byte) error {
				ns := namespaceFromKey(string(k))
				if ns != "" {
					seen[ns] = true
				}
				return nil
			})
		}); err != nil {
			return nil, err
		}
	}

	out := make([]string, 0, len(seen))
	for ns := range seen {
		out = append(out, ns)
	}
	sort.Strings(out)
	return out, nil
}
