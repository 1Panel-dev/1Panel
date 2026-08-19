package service

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"gorm.io/gorm"
)

type FileShareService struct{}

const (
	fileShareCodeMinLength     = 10
	fileShareCodeMaxLength     = 16
	fileShareCodeDefaultLength = 13
	fileShareCharset           = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

var fileShareCodeRegexp = regexp.MustCompile(`^[A-Za-z0-9]{10,16}$`)

type IFileShareService interface {
	Create(req request.FileShareCreate) (*response.FileShareInfo, error)
	Page(req dto.PageInfo, readOnly ...bool) (int64, []response.FileShareInfo, error)
	GetByPath(path string, readOnly ...bool) (*response.FileShareInfo, error)
	GetByCode(code string, readOnly ...bool) (*response.FileShareInfo, error)
	GetPublicByCode(code string, readOnly ...bool) (*response.FileSharePublicInfo, error)
	DeleteByPath(path string) error
	SharePathCodeMap() (map[string]string, error)
	Check(code, password string, readOnly ...bool) error
	PrepareDownload(code, password string, readOnly ...bool) (filePath, fileName string, err error)
}

func NewIFileShareService() IFileShareService {
	return &FileShareService{}
}

func randomShareCode(length int) (string, error) {
	if length < fileShareCodeMinLength || length > fileShareCodeMaxLength {
		length = fileShareCodeDefaultLength
	}
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	buf := make([]byte, length)
	for i := range b {
		buf[i] = fileShareCharset[int(b[i])%len(fileShareCharset)]
	}
	return string(buf), nil
}

func randomSalt() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashPassword(salt, password string) string {
	sum := sha256.Sum256([]byte(salt + ":" + password))
	return hex.EncodeToString(sum[:])
}

func shareModelToInfo(item model.FileShare) response.FileShareInfo {
	return response.FileShareInfo{
		Code:        item.Token,
		Path:        item.Path,
		FileName:    item.FileName,
		ExpiresAt:   item.ExpiresUnix,
		Permanent:   item.ExpiresUnix == 0,
		HasPassword: item.PasswordHash != "",
	}
}

func fillSharePassword(info *response.FileShareInfo, item model.FileShare) {
	if info == nil || item.PasswordEnc == "" {
		return
	}
	password, err := encrypt.StringDecrypt(item.PasswordEnc)
	if err != nil {
		return
	}
	info.Password = password
}

func shareModelToPublicInfo(item model.FileShare) response.FileSharePublicInfo {
	return response.FileSharePublicInfo{
		FileName:    item.FileName,
		ExpiresAt:   item.ExpiresUnix,
		Permanent:   item.ExpiresUnix == 0,
		HasPassword: item.PasswordHash != "",
	}
}

func (s *FileShareService) generateUniqueCode() (string, error) {
	for i := 0; i < 8; i++ {
		code, err := randomShareCode(fileShareCodeDefaultLength)
		if err != nil {
			return "", err
		}
		_, err = fileShareRepo.GetFirst(fileShareRepo.WithByCode(code))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code, nil
		}
		if err != nil {
			return "", err
		}
	}
	return "", errors.New("failed to generate unique file share code")
}

func (s *FileShareService) Create(req request.FileShareCreate) (*response.FileShareInfo, error) {
	path := strings.TrimSpace(req.Path)
	if path == "" || strings.Contains(path, "..") {
		return nil, buserr.New("ErrFileSharePath")
	}
	if files.ShouldDenySensitiveFileRead(path) {
		return nil, buserr.New("ErrSensitiveFileRead")
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return nil, buserr.New("ErrFileSharePath")
	}

	item, err := fileShareRepo.GetFirst(fileShareRepo.WithByPath(path))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	isNew := errors.Is(err, gorm.ErrRecordNotFound)
	if isNew {
		code, err := s.generateUniqueCode()
		if err != nil {
			return nil, err
		}
		item = model.FileShare{
			Path:     path,
			Token:    code,
			FileName: filepath.Base(path),
		}
	} else if !fileShareCodeRegexp.MatchString(item.Token) {
		code, err := s.generateUniqueCode()
		if err != nil {
			return nil, err
		}
		item.Token = code
	}

	item.FileName = filepath.Base(path)
	item.MaxDownloads = 0
	item.DownloadCount = 0
	item.ExpiresUnix = 0
	if req.ExpireMinutes > 0 {
		item.ExpiresUnix = time.Now().Add(time.Duration(req.ExpireMinutes) * time.Minute).Unix()
	}

	if req.Password != nil {
		pw := strings.TrimSpace(*req.Password)
		if pw == "" {
			item.PasswordEnc = ""
			item.PasswordSalt = ""
			item.PasswordHash = ""
		} else {
			pwLen := utf8.RuneCountInString(pw)
			if pwLen < 4 || pwLen > 256 {
				return nil, buserr.New("ErrFileSharePasswordPolicy")
			}
			enc, err := encrypt.StringEncrypt(pw)
			if err != nil {
				return nil, err
			}
			item.PasswordEnc = enc
			salt, err := randomSalt()
			if err != nil {
				return nil, err
			}
			item.PasswordSalt = salt
			item.PasswordHash = hashPassword(salt, pw)
		}
	}

	if isNew {
		if err := fileShareRepo.Create(&item); err != nil {
			return nil, err
		}
	} else {
		if err := fileShareRepo.Save(&item); err != nil {
			return nil, err
		}
	}

	res := shareModelToInfo(item)
	fillSharePassword(&res, item)
	return &res, nil
}

func (s *FileShareService) Page(req dto.PageInfo, readOnly ...bool) (int64, []response.FileShareInfo, error) {
	items, err := fileShareRepo.All()
	if err != nil {
		return 0, nil, err
	}
	result := make([]response.FileShareInfo, 0, len(items))
	for _, item := range items {
		if err := s.pruneInvalidShare(item, readOnly...); err != nil {
			return 0, nil, err
		}
		if item.ExpiresUnix > 0 && time.Now().Unix() > item.ExpiresUnix {
			continue
		}
		result = append(result, shareModelToInfo(item))
	}
	total := len(result)
	start := (req.Page - 1) * req.PageSize
	if start >= total {
		return int64(total), []response.FileShareInfo{}, nil
	}
	end := start + req.PageSize
	if end > total {
		end = total
	}
	return int64(total), result[start:end], nil
}

func (s *FileShareService) GetByPath(path string, readOnly ...bool) (*response.FileShareInfo, error) {
	item, err := fileShareRepo.GetFirst(fileShareRepo.WithByPath(strings.TrimSpace(path)))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if err := s.pruneInvalidShare(item, readOnly...); err != nil {
		return nil, err
	}
	if item.ExpiresUnix > 0 && time.Now().Unix() > item.ExpiresUnix {
		return nil, nil
	}
	info := shareModelToInfo(item)
	fillSharePassword(&info, item)
	return &info, nil
}

func (s *FileShareService) GetByCode(code string, readOnly ...bool) (*response.FileShareInfo, error) {
	item, err := fileShareRepo.GetFirst(fileShareRepo.WithByCode(strings.TrimSpace(code)))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, buserr.New("ErrFileShareInvalid")
		}
		return nil, err
	}
	if err := s.pruneInvalidShare(item, readOnly...); err != nil {
		return nil, err
	}
	if item.ExpiresUnix > 0 && time.Now().Unix() > item.ExpiresUnix {
		return nil, buserr.New("ErrFileShareExpired")
	}
	info := shareModelToInfo(item)
	return &info, nil
}

func (s *FileShareService) GetPublicByCode(code string, readOnly ...bool) (*response.FileSharePublicInfo, error) {
	item, err := fileShareRepo.GetFirst(fileShareRepo.WithByCode(strings.TrimSpace(code)))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, buserr.New("ErrFileShareInvalid")
		}
		return nil, err
	}
	if err := s.pruneInvalidShare(item, readOnly...); err != nil {
		return nil, err
	}
	if item.ExpiresUnix > 0 && time.Now().Unix() > item.ExpiresUnix {
		return nil, buserr.New("ErrFileShareExpired")
	}
	info := shareModelToPublicInfo(item)
	return &info, nil
}

func (s *FileShareService) DeleteByPath(path string) error {
	item, err := fileShareRepo.GetFirst(fileShareRepo.WithByPath(strings.TrimSpace(path)))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return buserr.New("ErrFileShareInvalid")
		}
		return err
	}
	return fileShareRepo.Delete(repo.WithByID(item.ID))
}

func (s *FileShareService) SharePathCodeMap() (map[string]string, error) {
	items, err := fileShareRepo.All()
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(items))
	now := time.Now().Unix()
	for _, item := range items {
		if item.ExpiresUnix > 0 && now > item.ExpiresUnix {
			continue
		}
		if _, err := os.Stat(item.Path); err != nil {
			continue
		}
		result[item.Path] = item.Token
	}
	return result, nil
}

func (s *FileShareService) Check(code, password string, readOnly ...bool) error {
	_, err := s.check(code, password, readOnly...)
	return err
}

func (s *FileShareService) PrepareDownload(code, password string, readOnly ...bool) (string, string, error) {
	item, err := s.check(code, password, readOnly...)
	if err != nil {
		return "", "", err
	}
	return item.Path, item.FileName, nil
}

func (s *FileShareService) pruneInvalidShare(item model.FileShare, readOnly ...bool) error {
	now := time.Now().Unix()
	if item.ExpiresUnix > 0 && now > item.ExpiresUnix {
		if isDemoReadOnly(readOnly...) {
			return nil
		}
		return fileShareRepo.Delete(repo.WithByID(item.ID))
	}
	info, err := os.Stat(item.Path)
	if err != nil || info.IsDir() {
		if isDemoReadOnly(readOnly...) {
			return nil
		}
		return fileShareRepo.Delete(repo.WithByID(item.ID))
	}
	return nil
}

func (s *FileShareService) check(code, password string, readOnly ...bool) (*model.FileShare, error) {
	code = strings.TrimSpace(code)
	password = strings.TrimSpace(password)
	if code == "" {
		return nil, buserr.New("ErrFileShareInvalid")
	}

	item, err := fileShareRepo.GetFirst(fileShareRepo.WithByCode(code))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, buserr.New("ErrFileShareInvalid")
		}
		return nil, err
	}

	now := time.Now().Unix()
	if item.ExpiresUnix > 0 && now > item.ExpiresUnix {
		if !isDemoReadOnly(readOnly...) {
			_ = fileShareRepo.Delete(repo.WithByID(item.ID))
		}
		return nil, buserr.New("ErrFileShareExpired")
	}
	if item.PasswordHash != "" {
		if subtle.ConstantTimeCompare([]byte(hashPassword(item.PasswordSalt, password)), []byte(item.PasswordHash)) != 1 {
			return nil, buserr.New("ErrFileSharePassword")
		}
	}

	info, err := os.Stat(item.Path)
	if err != nil || info.IsDir() {
		if !isDemoReadOnly(readOnly...) {
			_ = fileShareRepo.Delete(repo.WithByID(item.ID))
		}
		return nil, buserr.New("ErrFileSharePath")
	}

	return &item, nil
}
