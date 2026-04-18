const { envVersion } = wx.getAccountInfoSync().miniProgram

export const config = {
  apiBaseUrl: envVersion === 'release' || envVersion === 'trial'
    ? 'https://api.douxue.com'
    : 'http://192.168.1.19:3001',
}
