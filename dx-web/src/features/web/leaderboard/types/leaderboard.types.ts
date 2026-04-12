export type LeaderboardType = "exp" | "playtime";
export type LeaderboardPeriod = "day" | "week" | "month";

export type LeaderboardEntry = {
  id: string;
  username: string;
  nickname: string | null;
  avatarUrl: string | null;
  value: number;
  rank: number;
};

export type LeaderboardResult = {
  entries: LeaderboardEntry[];
  myRank: LeaderboardEntry | null;
};
