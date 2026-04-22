import { clearSessionAndRedirectToLogin } from './auth.js'

const DEFAULT_DEV_BASE_URL = 'http://127.0.0.1:8080/api/v1'
const DEFAULT_PROD_BASE_URL = '/api/v1'

function normalizeBaseURL(value) {
  const trimmed = String(value || '').trim().replace(/\/+$/, '')
  if (!trimmed) {
    return ''
  }
  if (trimmed.endsWith('/api/v1') || trimmed.endsWith('/v1')) {
    return trimmed
  }
  if (trimmed.endsWith('/api')) {
    return `${trimmed}/v1`
  }
  return trimmed
}

function resolveBaseURL() {
  const envBaseURL =
    typeof import.meta !== 'undefined' && import.meta.env
      ? normalizeBaseURL(import.meta.env.VITE_API_BASE_URL)
      : ''
  if (envBaseURL) {
    return envBaseURL
  }

  const isDev =
    typeof process !== 'undefined' &&
    process.env &&
    process.env.NODE_ENV === 'development'

  return isDev ? DEFAULT_DEV_BASE_URL : DEFAULT_PROD_BASE_URL
}

export const BASE_URL = resolveBaseURL()

export function request({ url, method = 'GET', data = {} }) {
  return new Promise((resolve, reject) => {
    const token = uni.getStorageSync('token');
    uni.request({
      url: BASE_URL + url,
      method,
      data,
      header: {
        'Authorization': token ? `Bearer ${token}` : '',
        'Content-Type': 'application/json'
      },
      success: (res) => {
        const result = res.data;
        if (res.statusCode === 200) {
          if (result.code === 0) {
            resolve(result.data);
          } else {
            uni.showToast({ title: result.message || '请求失败', icon: 'none' });
            if (result.code === 30004) {
              clearSessionAndRedirectToLogin();
            }
            reject(result);
          }
        } else {
          uni.showToast({ title: result.message || '网络连接异常', icon: 'none' });
          if (res.statusCode === 401) {
            clearSessionAndRedirectToLogin();
          }
          reject(result);
        }
      },
      fail: (err) => {
        uni.showToast({ title: '网络请求失败，请检查网络', icon: 'none' });
        reject(err);
      }
    });
  });
}

export function uploadFile({ url, filePath, name = 'file', formData = {} }) {
  return new Promise((resolve, reject) => {
    const token = uni.getStorageSync('token');
    uni.uploadFile({
      url: BASE_URL + url,
      filePath,
      name,
      formData,
      header: {
        'Authorization': token ? `Bearer ${token}` : ''
      },
      success: (res) => {
        if (res.statusCode === 200) {
          const result = JSON.parse(res.data);
          if (result.code === 0) {
            resolve(result.data);
          } else {
            uni.showToast({ title: result.message || '上传失败', icon: 'none' });
            if (result.code === 30004) {
              clearSessionAndRedirectToLogin();
            }
            reject(result);
          }
        } else {
          uni.showToast({ title: '上传异常', icon: 'none' });
          if (res.statusCode === 401) {
            clearSessionAndRedirectToLogin();
          }
          reject(res);
        }
      },
      fail: (err) => {
        uni.showToast({ title: '上传失败，检查网络', icon: 'none' });
        reject(err);
      }
    });
  });
}
