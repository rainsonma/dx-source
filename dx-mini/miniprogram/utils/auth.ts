const TOKEN_KEY = 'dx_token'
const USER_ID_KEY = 'dx_user_id'

export function getToken(): string | null {
  return (wx.getStorageSync(TOKEN_KEY) as string) || null
}
export function setToken(token: string): void {
  wx.setStorageSync(TOKEN_KEY, token)
}
export function clearToken(): void {
  wx.removeStorageSync(TOKEN_KEY)
  wx.removeStorageSync(USER_ID_KEY)
}
export function isLoggedIn(): boolean {
  return !!getToken()
}
export function getUserId(): string | null {
  return (wx.getStorageSync(USER_ID_KEY) as string) || null
}
export function setUserId(id: string): void {
  wx.setStorageSync(USER_ID_KEY, id)
}
