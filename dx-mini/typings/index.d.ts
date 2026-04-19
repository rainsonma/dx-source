/// <reference path="./types/index.d.ts" />

interface IAppOption {
  globalData: {
    userInfo?: WechatMiniprogram.UserInfo,
  }
  userInfoReadyCallback?: WechatMiniprogram.GetUserInfoSuccessCallback,
}

// Augment the bundled `miniprogram-api-typings@2.8.3-1` with post-2.x wx APIs
// that are present at runtime but absent from the pinned types.
declare namespace WechatMiniprogram {
  interface Wx {
    getSystemSetting(): {
      bluetoothEnabled?: boolean
      locationEnabled?: boolean
      wifiEnabled?: boolean
      deviceOrientation?: 'portrait' | 'landscape'
      theme?: 'light' | 'dark'
    }
    getAppBaseInfo(): {
      SDKVersion: string
      enableDebug?: boolean
      host?: { env: string }
      language?: string
      version: string
      theme?: 'light' | 'dark'
    }
    getDeviceInfo(): {
      brand: string
      model: string
      system: string
      platform: string
    }
    getWindowInfo(): {
      pixelRatio: number
      screenWidth: number
      screenHeight: number
      windowWidth: number
      windowHeight: number
      statusBarHeight: number
      safeArea: SafeArea
    }
  }
}