package service

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/spf13/afero"
)

type FavoriteService struct {
}

type IFavoriteService interface {
	Page(req dto.PageInfo) (int64, []response.FavoriteDTO, error)
	Create(req request.FavoriteCreate) (*model.Favorite, error)
	Delete(id uint) error
}

func NewIFavoriteService() IFavoriteService {
	return &FavoriteService{}
}

func (f *FavoriteService) Page(req dto.PageInfo) (int64, []response.FavoriteDTO, error) {
	total, favorites, err := favoriteRepo.Page(req.Page, req.PageSize)
	if err != nil {
		return 0, nil, err
	}
	var dtoFavorites []response.FavoriteDTO
	for _, favorite := range favorites {
		dtoFavorites = append(dtoFavorites, response.FavoriteDTO{
			Favorite: favorite,
		})
	}
	return total, dtoFavorites, nil
}

func (f *FavoriteService) Create(req request.FavoriteCreate) (*model.Favorite, error) {
	exist, _ := favoriteRepo.GetFirst(favoriteRepo.WithByPath(req.Path))
	if exist.ID > 0 {
		return nil, buserr.New("ErrFavoriteExist")
	}
	op := files.NewFileOp()
	if !op.Stat(req.Path) {
		return nil, buserr.New("ErrLinkPathNotFound")
	}
	openFile, err := op.OpenFile(req.Path)
	if err != nil {
		return nil, err
	}
	fileInfo, err := openFile.Stat()
	if err != nil {
		return nil, err
	}
	favorite := &model.Favorite{
		Name:  fileInfo.Name(),
		IsDir: fileInfo.IsDir(),
		Path:  req.Path,
	}
	if fileInfo.Size() <= 10*1024*1024 {
		afs := &afero.Afero{Fs: op.Fs}
		cByte, err := afs.ReadFile(req.Path)
		if err == nil {
			if len(cByte) > 0 && !files.DetectBinary(cByte) {
				favorite.IsTxt = true
			}
		}
	}
	if err := favoriteRepo.Create(favorite); err != nil {
		return nil, err
	}
	return favorite, nil
}

func (f *FavoriteService) Delete(id uint) error {
	if err := favoriteRepo.Delete(repo.WithByID(id)); err != nil {
		return err
	}
	return nil
}
