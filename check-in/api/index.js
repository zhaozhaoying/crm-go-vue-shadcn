import { request, uploadFile } from '../utils/request';

const GEO_API_URL = 'https://cn.apihz.cn/api/other/jwjuhe2.php'
const GEO_API_ID = '10014321'
const GEO_API_KEY = '02b0306e52a6aea3b64dae68396ecc60'

// 登录
export function login(data) {
  return request({ url: '/auth/login', method: 'POST', data });
}

// 验证码
export function getLoginCaptcha() {
  return request({ url: '/auth/captcha', method: 'GET' });
}

// 获取用户信息
export function getUserInfo() {
  return request({ url: '/auth/me', method: 'GET' });
}

// 签到上传图片
export function uploadVisitImg(filePath) {
  return uploadFile({
    url: '/users/avatar/upload',
    filePath,
    name: 'file'
  });
}

// 签到提交
export function createCustomerVisit(data) {
  return request({ url: '/customer-visits', method: 'POST', data });
}

// 获取签到记录
export function getCustomerVisits(params = {}) {
  let urlStr = '/customer-visits';
  const queryList = [];
  for (const key in params) {
    if (params[key] !== undefined && params[key] !== '') {
      queryList.push(`${key}=${encodeURIComponent(params[key])}`);
    }
  }
  if (queryList.length > 0) {
    urlStr += '?' + queryList.join('&');
  }
  return request({ url: urlStr, method: 'GET' });
}

// 获取系统配置
export function getSystemSettings() {
  return request({ url: '/settings', method: 'GET' });
}

export function reverseGeocodeByApihz({ lat, lon }) {
  return new Promise((resolve, reject) => {
    uni.request({
      url: GEO_API_URL,
      method: 'GET',
      data: {
        id: GEO_API_ID,
        key: GEO_API_KEY,
        lat,
        lon,
      },
      success: (res) => {
        const result = res.data || {}
        if (res.statusCode === 200 && Number(result.code) === 200) {
          resolve(result)
          return
        }
        reject(result)
      },
      fail: (err) => {
        reject(err)
      },
    })
  })
}
