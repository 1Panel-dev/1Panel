package repo

import "github.com/1Panel-dev/1Panel/agent/app/model"

type TensorRTLLMRepo struct {
}

type ITensorRTLLMRepo interface {
	Page(page, size int, opts ...DBOption) (int64, []model.TensorRTLLM, error)
	GetFirst(opts ...DBOption) (*model.TensorRTLLM, error)
	Create(tensorrtLLM *model.TensorRTLLM) error
	Save(tensorrtLLM *model.TensorRTLLM) error
	DeleteBy(opts ...DBOption) error
	List(opts ...DBOption) ([]model.TensorRTLLM, error)
}

func NewITensorRTLLMRepo() ITensorRTLLMRepo {
	return &TensorRTLLMRepo{}
}

func (t TensorRTLLMRepo) Page(page, size int, opts ...DBOption) (int64, []model.TensorRTLLM, error) {
	var servers []model.TensorRTLLM
	db := getDb(opts...).Model(&model.TensorRTLLM{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&servers).Error
	return count, servers, err
}

func (t TensorRTLLMRepo) GetFirst(opts ...DBOption) (*model.TensorRTLLM, error) {
	var tensorrtLLM model.TensorRTLLM
	if err := getDb(opts...).First(&tensorrtLLM).Error; err != nil {
		return nil, err
	}
	return &tensorrtLLM, nil
}

func (t TensorRTLLMRepo) List(opts ...DBOption) ([]model.TensorRTLLM, error) {
	var tensorrtLLMs []model.TensorRTLLM
	if err := getDb(opts...).Find(&tensorrtLLMs).Error; err != nil {
		return nil, err
	}
	return tensorrtLLMs, nil
}

func (t TensorRTLLMRepo) Create(tensorrtLLM *model.TensorRTLLM) error {
	return getDb().Create(tensorrtLLM).Error
}

func (t TensorRTLLMRepo) Save(tensorrtLLM *model.TensorRTLLM) error {
	return getDb().Save(tensorrtLLM).Error
}

func (t TensorRTLLMRepo) DeleteBy(opts ...DBOption) error {
	return getDb(opts...).Delete(&model.TensorRTLLM{}).Error
}
