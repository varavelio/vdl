import { error } from "@sveltejs/kit";

import type { PageLoad } from "./$types";

export type NodeKind = "type" | "proc" | "stream" | "doc" | "rpc";

export interface NodeRouteData {
  nodeSlug: string;
  nodeKind: NodeKind;
  nodeName: string;
  rpcName?: string;
}

export const load: PageLoad = async ({ params }): Promise<NodeRouteData> => {
  if (!params.nodeSlug) error(404, "Not found");

  const firstDashIndex = params.nodeSlug.indexOf("-");
  if (firstDashIndex === -1) error(404, "Not found");

  const nodeKindRaw = params.nodeSlug.substring(0, firstDashIndex);
  const remainder = params.nodeSlug.substring(firstDashIndex + 1);

  if (!nodeKindRaw || !remainder) error(404, "Not found");

  if (nodeKindRaw === "rpc") {
    const secondDashIndex = remainder.indexOf("-");
    if (secondDashIndex === -1) error(404, "Not found");

    const rpcName = remainder.substring(0, secondDashIndex);
    const nodeName = remainder.substring(secondDashIndex + 1);

    if (!rpcName || !nodeName) error(404, "Not found");

    return {
      nodeSlug: params.nodeSlug,
      nodeKind: "rpc",
      rpcName,
      nodeName,
    };
  }

  return {
    nodeSlug: params.nodeSlug,
    nodeKind: nodeKindRaw as NodeKind,
    nodeName: remainder,
  };
};
