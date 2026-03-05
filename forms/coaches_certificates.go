package forms

type CoachAddCertificatesRequest struct {
	CertificateReviews []CertificateReviews `json:"certificate_reviews"`
}

type CoachGetCertificatesRequest struct {
	CertificateIds []int64 `json:"certificate_ids" binding:"required"`
}

type CertificateReviews struct {
	Id                 int64    `json:"id"`
	CertificateId      int64    `json:"certificate_id" binding:"required"`
	Level              string   `json:"level" binding:"required"`
	CertificateImgUrls []string `json:"certificate_img_urls" binding:"required"`
}
