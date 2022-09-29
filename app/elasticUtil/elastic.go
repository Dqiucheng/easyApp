package elasticUtil

import (
	"context"
	"easyApp/db"
	"github.com/olivere/elastic/v7"
)

type Elastic struct {
	Ctx context.Context
}

// KeywordAnalyzerPreview 关键词分词效果预览
func (receiver Elastic) KeywordAnalyzerPreview(analyzer, text string) (*elastic.IndicesAnalyzeResponse, error) {
	return db.Es().IndexAnalyze().Analyzer(analyzer).Text(text).Do(receiver.Ctx)
}

// IndexExists 判断索引是否存在
func (receiver Elastic) IndexExists(indexName string) (bool, error) {
	return db.Es().IndexExists(indexName).Do(receiver.Ctx)
}

// DeleteIndex 删除索引
func (receiver Elastic) DeleteIndex(indexName string) (*elastic.IndicesDeleteResponse, error) {
	return db.Es().DeleteIndex(indexName).IgnoreUnavailable(true).Do(receiver.Ctx)
}

// CreateIndex 创建索引
func (receiver Elastic) CreateIndex(indexName string, mapping interface{}) (*elastic.IndicesCreateResult, error) {
	return db.Es().CreateIndex(indexName).BodyJson(mapping).Do(receiver.Ctx)
}

// GetMapping 查看索引映射
// indexName	参数可传空
func (receiver *Elastic) GetMapping(indexName string) (map[string]interface{}, error) {
	return db.Es().GetMapping().Index(indexName).Do(receiver.Ctx)
}

// PutMapping 更新索引映射
func (receiver Elastic) PutMapping(indexName string, mapping map[string]interface{}) (*elastic.PutMappingResponse, error) {
	return db.Es().PutMapping().Index(indexName).BodyJson(mapping).Do(receiver.Ctx)
}

// CreateDocToIndex 创建文档
func (receiver Elastic) CreateDocToIndex(indexName string, doc map[string]interface{}) (*elastic.IndexResponse, error) {
	return db.Es().Index().Index(indexName).BodyJson(doc).Do(receiver.Ctx)
}

// UpdateOrCreateIdDoc 创建或更新文档
func (receiver Elastic) UpdateOrCreateIdDoc(indexName, docId string, doc map[string]interface{}) (*elastic.UpdateResponse, error) {
	return db.Es().Update().
		Index(indexName).
		Id(docId).
		Doc(doc).
		DocAsUpsert(true). // 如果文档不存在则插入
		Do(receiver.Ctx)
}

// DeleteIdDoc 删除文档
func (receiver Elastic) DeleteIdDoc(indexName, docId string) (*elastic.DeleteResponse, error) {
	return db.Es().Delete().
		Index(indexName).
		Id(docId).
		Do(receiver.Ctx)
}

// GetIdDoc 按id查询单个文档
func (receiver Elastic) GetIdDoc(indexName string, docId string) (*elastic.GetResult, error) {
	return db.Es().Get().Index(indexName).Id(docId).Do(receiver.Ctx)
}

// BoolQuery 按条件查询文档
func (receiver Elastic) BoolQuery(indexName, name string, val interface{}) (*elastic.SearchResult, error) {
	// 创建新的bool查询
	query := elastic.NewBoolQuery()

	// must
	query.Must(elastic.NewTermQuery(name, val))

	//source, _ := query.Source()
	//fmt.Println(source)	// 打印bool查询语句

	return db.Es().Search().Index(indexName).Query(query).Do(receiver.Ctx)
}
