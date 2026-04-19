const DEFAULT_DEV_API_BASE_URL = 'http://192.168.1.19'
const DEV_URL_STORAGE_KEY = 'dx_dev_api_base_url'

const { envVersion } = wx.getAccountInfoSync().miniProgram
const isProdLike = envVersion === 'release' || envVersion === 'trial'

function resolveDevApiBaseUrl(): string {
  try {
    const stored = wx.getStorageSync(DEV_URL_STORAGE_KEY) as unknown
    if (typeof stored === 'string' && stored.trim() !== '') {
      return stored.trim()
    }
  } catch {
    // fall through
  }
  return DEFAULT_DEV_API_BASE_URL
}

export const config = {
  get apiBaseUrl(): string {
    return isProdLike ? 'https://api.douxue.com' : resolveDevApiBaseUrl()
  },
}

// Override the dev API URL. Run from DevTools console:
//   require('./utils/config').setDevApiBaseUrl('http://192.168.1.100')
// then reload the mini program.
export function setDevApiBaseUrl(url: string): void {
  wx.setStorageSync(DEV_URL_STORAGE_KEY, url)
}

export function clearDevApiBaseUrl(): void {
  wx.removeStorageSync(DEV_URL_STORAGE_KEY)
}
