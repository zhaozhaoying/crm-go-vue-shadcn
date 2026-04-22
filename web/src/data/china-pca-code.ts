import rawChinaPcaCode from "./china-pca-code.json"

export interface ChinaPcaNode {
  code: string
  name: string
  children?: ChinaPcaNode[]
}

// Source: https://github.com/modood/Administrative-divisions-of-China (dist/pca-code.json)
export const chinaPcaCode = rawChinaPcaCode as ChinaPcaNode[]
