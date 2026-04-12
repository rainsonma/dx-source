import { z } from "zod";

export const leaderboardParamsSchema = z.object({
  type: z.enum(["exp", "playtime"]),
  period: z.enum(["day", "week", "month"]),
});

export type LeaderboardParams = z.infer<typeof leaderboardParamsSchema>;
