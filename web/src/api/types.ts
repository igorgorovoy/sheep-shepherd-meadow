// TypeScript interfaces mirroring the Shepherd REST API JSON shapes.
// memory is in bytes, cpu is in millicores.

export interface ObjectMeta {
  name: string
  namespace: string
  uid: string
  labels?: Record<string, string>
  created_at: string
}

export interface ResourceList {
  cpu: number
  memory: number
  pods: number
}

// ---- Node ----------------------------------------------------------------

export type NodeCondition = 'Ready' | 'NotReady'

export interface NodeSpec {
  address: string
}

export interface NodeStatus {
  condition: NodeCondition
  capacity: ResourceList
  allocatable: ResourceList
  pod_count: number
  last_heartbeat: string
}

export interface Node {
  kind: string
  metadata: ObjectMeta
  spec: NodeSpec
  status: NodeStatus
}

// ---- Pod -----------------------------------------------------------------

export type PodPhase = 'Pending' | 'Running' | 'Succeeded' | 'Failed'

export interface ContainerPort {
  name?: string
  container_port?: number
  protocol?: string
}

export interface ContainerResources {
  cpu?: number
  memory?: number
}

export interface Container {
  name: string
  image: string
  command?: string[]
  env?: Record<string, string>
  ports?: ContainerPort[]
  resources?: ContainerResources
}

export interface PodSpec {
  containers: Container[]
  node_name?: string
  restart_policy: string
  node_selector?: Record<string, string>
}

export interface ContainerStatus {
  name: string
  container_id: string
  ready: boolean
  state: string
  exit_code?: number
}

export interface PodStatus {
  phase: PodPhase
  host_ip?: string
  pod_ip?: string
  start_time?: string
  containers?: ContainerStatus[]
  message?: string
}

export interface Pod {
  kind: string
  metadata: ObjectMeta
  spec: PodSpec
  status: PodStatus
}

// ---- Service -------------------------------------------------------------

export type ServiceType = 'ClusterIP' | 'NodePort'

export interface ServicePort {
  name?: string
  port: number
  target_port: number
  node_port?: number
  protocol?: string
}

export interface ServiceSpec {
  selector: Record<string, string>
  ports: ServicePort[]
  type: ServiceType
}

export interface ServiceStatus {
  cluster_ip?: string
  endpoints?: string[]
}

export interface Service {
  kind: string
  metadata: ObjectMeta
  spec: ServiceSpec
  status: ServiceStatus
}

// ---- Deployment ----------------------------------------------------------

export interface DeploymentSpec {
  replicas: number
  selector: Record<string, string>
  template: unknown
}

export interface DeploymentStatus {
  replicas: number
  ready_replicas: number
  available_replicas: number
  updated_replicas: number
}

export interface Deployment {
  kind: string
  metadata: ObjectMeta
  spec: DeploymentSpec
  status: DeploymentStatus
}

// ---- Event ---------------------------------------------------------------

export type EventType = 'Normal' | 'Warning'

export interface Event {
  type: EventType
  reason: string
  message: string
  object: string
  timestamp: string
}

// ---- Info ----------------------------------------------------------------

export interface Info {
  version: string
  name: string
  node_count: number
  pod_count: number
}

// ---- Optional aggregate --------------------------------------------------

export interface ClusterSummary {
  info?: Info
  nodes?: Node[]
  pods?: Pod[]
  services?: Service[]
  deployments?: Deployment[]
  events?: Event[]
}
