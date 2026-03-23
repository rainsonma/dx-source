import { z } from "zod";

export const leaderboardParamsSchema = z.object({
  type: z.enum(["exp", "playTime"]),
  period: z.enum(["all", "day", "week", "month"]),
});

export type LeaderboardParams = z.infer<typeof leaderboardParamsSchema>;
