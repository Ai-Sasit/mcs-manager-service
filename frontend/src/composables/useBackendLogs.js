import { WS_BASE_URL } from "@/constants";
import { getToken } from "@/utils/authStorage";
import { useResumableLogStream } from "./useResumableLogStream";

export function useBackendLogs() {
  return useResumableLogStream(
    () => `${WS_BASE_URL}/ws/backend-logs?token=${encodeURIComponent(getToken() || "")}`,
    ["log"],
  );
}
