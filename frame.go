package jewl

type FrameIndex = int

type Frame struct{
    Location    string          `json:"location"`
    Name        string          `json:"name"`
    Args        map[string]any  `json:"args"`
    Start       int             `json:"start"`
    End         int             `json:"end"`
    Duration    int             `json:"duration"`
    Calls       []FrameIndex    `json:"calls"`
    Subframes   []FrameIndex    `json:"subframes"`
}
