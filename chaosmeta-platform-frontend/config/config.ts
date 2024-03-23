import { defineConfig } from '@umijs/max';
import routes from './router';

export default defineConfig({
  title: 'chaosmeta',
  antd: {},
  access: {},
  model: {},
  esbuildMinifyIIFE: true,
  initialState: {},
  request: {},
  layout: {
    title: 'chaosmeta',
    locale: true
  },
  hash: true,
  historyWithQuery: {},
  routes,
  npmClient: 'yarn',
  styledComponents: {},
  locale: {
    // 默认使用 src/locales/zh-CN.ts 作为多语言文件
    default: 'zh-CN',
    baseSeparator: '-',
    baseNavigator: true,
    antd: true,
  },
  proxy: {
    '/users': {
      target: 'http://47.117.167.206:8082/',
      changeOrigin: true,
      pathRewrite: { '^/api': '' },
    },
    '/chaosmeta': {
      target: 'http://47.117.167.206:8082/',
      changeOrigin: true,
      pathRewrite: { '^/api': '' },
    },
  },
});
