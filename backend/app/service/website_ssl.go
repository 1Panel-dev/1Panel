package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
)

type WebSiteSSLService struct {
}

func (w WebSiteSSLService) Page(search dto.PageInfo) (int64, []dto.WebsiteSSLDTO, error) {
	total, sslList, err := websiteSSLRepo.Page(search.Page, search.PageSize, commonRepo.WithOrderBy("created_at desc"))
	var sslDTOs []dto.WebsiteSSLDTO
	for _, ssl := range sslList {
		sslDTOs = append(sslDTOs, dto.WebsiteSSLDTO{
			WebSiteSSL: ssl,
		})
	}
	return total, sslDTOs, err
}

//func (w WebSiteSSLService) Create(create dto.WebsiteSSLCreate) (dto.WebsiteSSLCreate, error) {
//
//	authorization, err := json.Marshal(create.Authorization)
//	if err != nil {
//		return dto.WebsiteSSLCreate{}, err
//	}
//
//	if err := websiteSSLRepo.Create(model.WebsiteDnsAccount{
//		Name:          create.Name,
//		Type:          create.Type,
//		Authorization: string(authorization),
//	}); err != nil {
//		return dto.WebsiteSSLCreate{}, err
//	}
//
//	return create, nil
//}
//
//func (w WebSiteSSLService) Update(update dto.WebsiteDnsAccountUpdate) (dto.WebsiteDnsAccountUpdate, error) {
//
//	authorization, err := json.Marshal(update.Authorization)
//	if err != nil {
//		return dto.WebsiteDnsAccountUpdate{}, err
//	}
//
//	if err := websiteSSLRepo.Save(model.WebsiteDnsAccount{
//		BaseModel: model.BaseModel{
//			ID: update.ID,
//		},
//		Name:          update.Name,
//		Type:          update.Type,
//		Authorization: string(authorization),
//	}); err != nil {
//		return dto.WebsiteDnsAccountUpdate{}, err
//	}
//
//	return update, nil
//}

func (w WebSiteSSLService) Delete(id uint) error {
	return websiteSSLRepo.DeleteBy(commonRepo.WithByID(id))
}
