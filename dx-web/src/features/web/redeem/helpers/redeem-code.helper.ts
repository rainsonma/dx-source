const CHARSET = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
const GROUP_SIZE = 4;
const GROUP_COUNT = 4;

/** Generate a cryptographically random redeem code in XXXX-XXXX-XXXX-XXXX format */
export function generateRedeemCode(): string {
  const totalChars = GROUP_SIZE * GROUP_COUNT;
  const bytes = new Uint8Array(totalChars);
  crypto.getRandomValues(bytes);

  const chars = Array.from(bytes, (b) => CHARSET[b % CHARSET.length]);
  const groups: string[] = [];

  for (let i = 0; i < GROUP_COUNT; i++) {
    groups.push(chars.slice(i * GROUP_SIZE, (i + 1) * GROUP_SIZE).join(""));
  }

  return groups.join("-");
}
