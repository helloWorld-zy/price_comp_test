package prompts

// QuoteParsePrompt generates the prompt for parsing price quotes from text
func QuoteParsePrompt(text string) string {
	return `你是一个邮轮航次报价信息提取专家。请从以下文本中提取报价信息，以JSON格式返回。

要提取的字段：
- sailing_code: 航次编号
- ship_name: 邮轮名称
- departure_date: 出发日期 (YYYY-MM-DD)
- nights: 晚数
- route: 航线
- quotes: 报价列表，每个报价包含:
  - cabin_type_name: 房型名称
  - cabin_category: 房型大类 (内舱/海景/阳台/套房)
  - price: 价格
  - currency: 币种
  - pricing_unit: 计价口径 (PER_PERSON/PER_CABIN/TOTAL)
  - conditions: 适用条件
  - promotion: 促销信息
  - notes: 备注

文本内容：
` + text + `

请以JSON格式返回结果。`
}
