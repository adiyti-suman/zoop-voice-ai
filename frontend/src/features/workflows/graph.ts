/**
 * graph.ts
 * Converts the backend WorkflowGraph model into React Flow–compatible
 * `nodes` and `edges` arrays, ready to drop into a <ReactFlow /> component.
 */
import type { WorkflowNode, WorkflowEdge } from "./types";

// ── React Flow node shape ─────────────────────────────────────────────────────

export interface RFNode {
  id: string;
  type: "default" | "input" | "output";
  position: { x: number; y: number };
  data: {
    label: string;
    nodeType: string;
    config: Record<string, unknown>;
    nodeKey: string;
  };
}

export interface RFEdge {
  id: string;
  source: string; // node_key of source
  target: string; // node_key of target
  label?: string;
  animated?: boolean;
}

// ── Converters ────────────────────────────────────────────────────────────────

/**
 * Convert a backend WorkflowNode into a React Flow node.
 * Uses the node_key as the stable React Flow id so that edges resolve correctly.
 */
export function toRFNode(node: WorkflowNode): RFNode {
  let rfType: "default" | "input" | "output" = "default";
  if (node.type === "START") rfType = "input";
  if (node.type === "END") rfType = "output";

  return {
    id: node.node_key,
    type: rfType,
    position: node.position ?? { x: 0, y: 0 },
    data: {
      label: node.name || node.node_key,
      nodeType: node.type,
      config: node.config ?? {},
      nodeKey: node.node_key,
    },
  };
}

/**
 * Convert a backend WorkflowEdge into a React Flow edge.
 */
export function toRFEdge(edge: WorkflowEdge): RFEdge {
  return {
    id: edge.id,
    source: edge.source_node_key,
    target: edge.target_node_key,
    label: edge.condition || undefined,
    animated: !!edge.condition,
  };
}

/**
 * Convert an entire backend graph into React Flow format.
 */
export function toReactFlow(
  nodes: WorkflowNode[],
  edges: WorkflowEdge[]
): { nodes: RFNode[]; edges: RFEdge[] } {
  return {
    nodes: nodes.map(toRFNode),
    edges: edges.map(toRFEdge),
  };
}
