import { WS_BASE_URL } from "@/constants";
import { getToken } from "@/utils/authStorage";
import { useResumableLogStream } from "./useResumableLogStream";

export function useServerLogs(serverId) {
  return useResumableLogStream(
    () =>
      `${WS_BASE_URL}/ws/servers/${encodeURIComponent(serverId)}/logs?token=${encodeURIComponent(getToken() || "")}`,
    ["log"],
  );
}
