package S3Service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	localConfig "hands-on-aws/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/samber/lo"
)

var AcceptedFileExtensions = []string{
	".jpg",
	".jpeg",
	".png",
	".bmp",
	".webp",
	".gif",
	".tif",
	".tiff",
	".apng",
	".pdf",
	".doc",
	".docx",
	".xls",
	".xlsx",
	".ppt",
	".pptx",
	".mov",
	".wmv",
	".mkv",
	".mp4",
	".webm",
	".avi",
	".mpeg",
	".mpg",
	".mp3",
	".wav",
	".wma",
	".aac",
	".flac",
	".m4a",
	".ogg",
	".opus",
	".zip",
	".rar",
	".tar",
	".gz",
	".7z",
	".json",
	".xml",
	".csv",
	".tsv",
	".txt",
	".log",
	".md",
	".html",
	".htm",
	".css",
	".js",
	".jsx",
	".ts",
	".tsx",
}

type S3Service struct {
	Bucket                 string
	s3Client               *s3.Client
	cfg                    aws.Config
	ctx                    context.Context
	PreSignedUrlExp        time.Duration
	AcceptedFileExtensions []string
	BucketIsPrivate        bool
	Region                 string
}

func NewForPublicBucket(ctx context.Context) *S3Service {
	accessKeyID := localConfig.Params.GetString("aws.key")
	secretAccessKey := localConfig.Params.GetString("aws.secret")
	sessionToken := localConfig.Params.GetString("aws.session")
	bucket := localConfig.Params.GetString("aws.bucket")
	region := localConfig.Params.GetString("aws.region")

	var cfg aws.Config
	cfg, _ = config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, sessionToken),
		),
	)

	return &S3Service{
		Bucket:                 bucket,
		s3Client:               s3.NewFromConfig(cfg),
		cfg:                    cfg,
		ctx:                    ctx,
		PreSignedUrlExp:        0,
		AcceptedFileExtensions: AcceptedFileExtensions,
		BucketIsPrivate:        false,
		Region:                 region,
	}
}

// SetPreSignedUrlExp : Önceden imzalanmış URL'nin geçerlilik süresini tekrar ayarlar
func (s *S3Service) SetPreSignedUrlExp(expiration time.Duration) {
	s.PreSignedUrlExp = expiration
}

// GetMimeTypeFromFileExtension : Verilen dosya uzantısına göre MIME türünü döner
//
// Usage:
//
//	s3Service := S3ServiceV2.NewForPublicBuket(ctx)
//	contentType := s3Service.GetMimeTypeFromFileExtension(".jpg")
//	fmt.Println("Content Type: ", contentType)
//
// Parameters:
// - fileExtension : string - Dosya uzantısı (örneğin, ".jpg", ".png")
//
// Returns:
// - contentType : string - Dosya uzantısına karşılık gelen MIME türü
func (s *S3Service) GetMimeTypeFromFileExtension(fileExtension string) (contentType string) {
	switch fileExtension {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".bmp":
		contentType = "image/bmp"
	case ".webp":
		contentType = "image/webp"
	case ".gif":
		contentType = "image/gif"
	case ".tif", ".tiff":
		contentType = "image/tiff"
	case ".apng":
		contentType = "image/apng"
	case ".pdf":
		contentType = "application/pdf"
	case ".doc":
		contentType = "application/msword"
	case ".docx":
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		contentType = "application/vnd.ms-excel"
	case ".xlsx":
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		contentType = "application/vnd.ms-powerpoint"
	case ".pptx":
		contentType = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".mov":
		contentType = "video/quicktime"
	case ".wmv":
		contentType = "video/x-ms-wmv"
	case ".mkv":
		contentType = "video/x-matroska"
	case ".mp4":
		contentType = "video/mp4"
	case ".webm":
		contentType = "video/webm"
	case ".avi":
		contentType = "video/x-msvideo"
	case ".mpeg", ".mpg":
		contentType = "video/mpeg"
	case ".mp3":
		contentType = "audio/mpeg"
	case ".wav":
		contentType = "audio/wav"
	case ".wma":
		contentType = "audio/x-ms-wma"
	case ".aac":
		contentType = "audio/aac"
	case ".flac":
		contentType = "audio/flac"
	case ".m4a":
		contentType = "audio/mp4"
	case ".ogg":
		contentType = "audio/ogg"
	case ".opus":
		contentType = "audio/opus"
	case ".zip":
		contentType = "application/zip"
	case ".rar":
		contentType = "application/vnd.rar"
	case ".tar":
		contentType = "application/x-tar"
	case ".gz":
		contentType = "application/gzip"
	case ".7z":
		contentType = "application/x-7z-compressed"
	case ".json":
		contentType = "application/json"
	case ".xml":
		contentType = "application/xml"
	case ".csv":
		contentType = "text/csv"
	case ".tsv":
		contentType = "text/tab-separated-values"
	case ".txt":
		contentType = "text/plain"
	case ".log":
		contentType = "text/plain"
	case ".md":
		contentType = "text/markdown"
	case ".html", ".htm":
		contentType = "text/html"
	case ".css":
		contentType = "text/css"
	case ".js":
		contentType = "application/javascript"
	case ".jsx":
		contentType = "text/jsx"
	case ".ts":
		contentType = "application/typescript"
	case ".tsx":
		contentType = "text/tsx"
	default:
		contentType = "application/octet-stream"
	}
	return
}

// Upload : Verilen veriyi S3'e yükler ve dosya yolunu döner
//
// Usage:
//
//	ctx := context.Background()
//	s3Service := S3ServiceV2.NewForPublicBuket(ctx)
//	filePath, err := s3Service.Upload("directory", "name", data)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println("File Path: ", filePath)
//
// Parameters:
// - directory : string - Dosyanın yükleneceği dizin
// - name : string - Dosya adı
// - data : []byte - Yüklenecek dosya verisi
//
// Returns:
// - filePath : string - Yüklenen dosyanın yolu
// - err : error - Yükleme sırasında oluşan hata, eğer varsa
func (s *S3Service) Upload(filePrefix, fileExtension string, data []byte) (filePath string, err error) {

	// Dosya boyutu kontrolü
	if s.CheckFileSize(int64(len(data))) {
		err = errors.New("file size is bigger than 50 MB")
		return
	}

	var contentType string
	contentType = http.DetectContentType(data)
	if contentType == "application/octet-stream" || contentType == "text/plain; charset=utf-8" {
		contentType = s.GetMimeTypeFromFileExtension(fileExtension)
	}

	// Dosya uzantı kontrolü
	if !lo.Contains(s.AcceptedFileExtensions, strings.ToLower(fileExtension)) {
		err = errors.New("file extension is not valid")
		return
	}

	// S3 yükleme
	uploadFile := &s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(filePrefix),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}

	_, err = s.s3Client.PutObject(s.ctx, uploadFile)
	if err != nil {
		return
	}

	filePath = fmt.Sprintf(
		"https://%s.s3.%s.amazonaws.com/%s",
		s.Bucket,
		s.Region,
		filePrefix,
	)

	return
}

// CheckFileSize : Verilen dosya boyutunun maksimum sınırı aşıp aşmadığını kontrol eder
//
// Usage:
//
//	s3Service := S3ServiceV2.NewForPublicBuket(ctx)
//	isValid := s3Service.CheckFileSize(fileSize)
//	if !isValid {
//		log.Println("File size exceeds the maximum limit.")
//	}
//
// Parametreler:
// - fileSize : int64 - Kontrol edilecek dosya boyutu (byte cinsinden)
//
// Dönüş:
// - bool : Dosya boyutu maksimum sınırı aşmıyorsa true, aksi halde false döner
func (s *S3Service) CheckFileSize(fileSize int64) bool {
	maxSize := int64(150 * 1024 * 1024) // 50 MB in bytes

	if fileSize > maxSize {
		return false
	} else {
		return true
	}
}

// DeleteObject : S3'teki bir nesneyi siler
//
// Usage:
//
//	ctx := context.Background()
//	s3Service := S3ServiceV2.NewForPublicBuket(ctx)
//	err := s3Service.DeleteObject("your-key")
//	if err != nil {
//		log.Fatal(err)
//	}
func (s *S3Service) DeleteObject(key string) (err error) {

	_, err = s.s3Client.DeleteObject(s.ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	})

	return
}
