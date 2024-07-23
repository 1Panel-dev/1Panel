package request

type RecycleBinCreate struct {
	SourcePath string `json:"sourcePath" validate:"required"`
}

type RecycleBinReduce struct {
	From  string `json:"from" validate:"required"`
	RName string `json:"rName" validate:"required"`
	Name  string `json:"name"`
}
