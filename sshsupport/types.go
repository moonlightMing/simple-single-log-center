package sshsupport

const (
    FILE_TYPE   = iota
    FOLDER_TYPE
    LINK_TYPE
    OTHER_TYPE
)

type UnixFile struct {
    Name  string `json:"name"`
    Type  int    `json:"type"` // file or folder
    Size  int64  `json:"size"`
    Group string `json:"group"`
    Owner string `json:"owner"`
}
