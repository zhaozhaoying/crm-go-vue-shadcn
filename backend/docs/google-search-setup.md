# Google Search Setup

Google 搜索使用 Google Custom Search JSON API。

## 必填配置

- `GOOGLE_API_KEY`: Google Cloud 中启用 Custom Search JSON API 后生成的 API Key
- `GOOGLE_CX`: Programmable Search Engine 的搜索引擎 ID
- `GOOGLE_SEARCH_NUM`: 每页抓取数量，最大 `10`

`.env` 示例：

```env
GOOGLE_API_KEY=your_google_api_key
GOOGLE_CX=your_search_engine_cx
GOOGLE_SEARCH_NUM=10
```

## 受限网络环境

如果后端所在机器无法直接访问 `https://www.googleapis.com`，Google 任务会失败。

这时需要配置代理：

```env
GOOGLE_PROXY_URL=http://127.0.0.1:7890
```

也可以在启动服务前设置：

```bash
export HTTPS_PROXY=http://127.0.0.1:7890
```

优先推荐 `GOOGLE_PROXY_URL`，因为它只影响 Google provider，不会改变 Alibaba 和 Made-in-China 的请求路径。

## 联调检查

运行下面的检查命令：

```bash
go run ./cmd/google-search-check --keyword "led light manufacturer"
```

成功时会输出抓取到的前几条结果；失败时会直接打印 Google API 返回信息，或者提示当前机器无法访问 Google 并建议配置代理。
