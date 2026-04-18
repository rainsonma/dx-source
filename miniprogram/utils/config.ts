const { envVersion } = wx.getAccountInfoSync().miniProgram

export const config = {
  apiBaseUrl: envVersion === 'release' || envVersion === 'trial'
    ? 'https://api.douxue.com'
    : 'http://localhost:3001',
}
