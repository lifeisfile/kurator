package lib

type TaskListItem struct {
	TaskTitle      string `json:"taskTitle"`
	TaskID         string `json:"taskID"`
	IsCompleted    bool   `json:"isCompleted"`
	IsLocked       bool   `json:"isLocked"`
	PositionNumber int    `json:"positionNumber"`
}

type CourseInfo struct {
	CourseTitle          string         `json:"courseTitle" yaml:"courseTitle"`
	CourseIcon           string         `json:"courseIcon" yaml:"courseIcon"`
	CourseSource         string         `json:"-" yaml:"courseSource"`
	AuthorName           string         `json:"authorName" yaml:"authorName"`
	AuthorPosition       string         `json:"authorPosition" yaml:"authorPosition"`
	AuthorPhoto          string         `json:"authorPhoto" yaml:"authorPhoto"`
	BaseServerHandlerURL string         `json:"-" yaml:"baseServerHandlerURL"`
	TaskList             []TaskListItem `json:"taskList"`
}

type TaskShortInfo struct {
	TaskTitle string `json:"taskTitle" yaml:"taskTitle"`
	TaskID    string `json:"taskID" yaml:"taskID"`
	IsFree    bool   `json:"isFree" yaml:"isFree"`
}

type TaskInfo struct {
	TaskTitle string   `json:"taskTitle" yaml:"taskTitle"`
	TaskID    string   `json:"taskID" yaml:"taskID"`
	IsFree    bool     `json:"isFree" yaml:"isFree"`
	Intro     string   `json:"intro" yaml:"intro"`
	DependsOn []string `json:"dependsOn" yaml:"dependsOn"`
	Goals     []struct {
		ID            string `json:"id" yaml:"id"`
		StatusHandler string `json:"statusHandler" yaml:"statusHandler"`
		RunHandler    string `json:"runHandler" yaml:"runHandler"`
		IsCompleted   bool   `json:"isCompleted"`
		Contents      []struct {
			Kind string `json:"kind" yaml:"kind"`
			Tabs []struct {
				Title   string `json:"title" yaml:"title"`
				Content string `json:"content" yaml:"content"`
			} `json:"tabs,omitempty" yaml:"tabs,omitempty"`
			Content        string `json:"content,omitempty" yaml:"content,omitempty"`
			SourceHandler  string `json:"sourceHandler,omitempty" yaml:"sourceHandler,omitempty"`
			CacheKey       string `json:"cacheKey,omitempty" yaml:"cacheKey,omitempty"`
			ID             string `json:"id,omitempty" yaml:"id,omitempty"`
			Title          string `json:"title,omitempty" yaml:"title,omitempty"`
			Text           string `json:"text,omitempty" yaml:"text,omitempty"`
			DisableOnClick int    `json:"disableOnClick,omitempty" yaml:"disableOnClick,omitempty"`
			AllowMultiple  bool   `json:"allowMultiple,omitempty" yaml:"allowMultiple,omitempty"`
			Answers        []struct {
				Text      string `json:"text,omitempty" yaml:"text,omitempty"`
				IsCorrect bool   `json:"isCorrect,omitempty" yaml:"isCorrect,omitempty"`
			} `json:"answers" yaml:"answers"`
			KuratorRequest struct {
				APIVersion string   `yaml:"apiVersion"`
				Type       string   `yaml:"type"`
				Payload    string   `yaml:"payload"`
				Command    string   `yaml:"command"`
				Args       []string `yaml:"args"`
				Files      []string `yaml:"files"`
			} `yaml:"kuratorRequest"`
		} `json:"contents" yaml:"contents"`
	} `json:"goals" yaml:"goals"`
	Faqs []struct {
		Question string `json:"question" yaml:"question"`
		Answer   string `json:"answer" yaml:"answer"`
	} `json:"faqs" yaml:"faqs"`
	RelatedCourses []string `json:"relatedCourses" yaml:"relatedCourses"`
}

type RequestHandler struct {
	Method                 string            `json:"method"`
	CourseName             string            `json:"courseName"`
	TaskNumber             int               `json:"taskNumber"`
	CacheKey               string            `json:"cacheKey"`
	KuratorCommandOutput   string            `json:"kuratorCommandOutput"`   // sets based on local command execution
	KuratorCommandExitCode int               `json:"kuratorCommandExitCode"` // sets based on local command execution
	BoolResponse           bool              `json:"boolResponse"`
	FilesBase64            map[string]string `json:"filesBase64"`
	UserID                 int64             `json:"userID"` // sets internally
	IsPaid                 bool              `json:"isPaid"` // sets internally
}

type KuratorRequest struct {
	ApiVersion string   `json:"apiVersion"`
	Type       string   `json:"type"`
	Payload    string   `json:"payload"`
	Command    string   `json:"command"`
	Args       []string `json:"args"`
	Files      []string `json:"files"`
	UserID     int64    `json:"user_id"`
	SeqID      int64    `json:"seq_id"`
	CourseName string   `json:"course_name"`
	IsDev      bool     `json:"is_dev"`
}

type KuratorResponse struct {
	CommandOutput   string            `json:"commandOutput"`
	CommandExitCode int               `json:"commandExitCode"`
	BoolResponse    bool              `json:"boolResponse"`
	FilesBase64     map[string]string `json:"filesBase64"`
	SeqID           int64             `json:"seq_id"`
}

type ResponseFromHandler struct {
	Status string
}

type ResponseEmpty struct {
}

type GenTask struct {
	TaskID  string
	IsFree  bool
	Methods []struct {
		HandlerName     string
		HandlerFuncName string
		IsOutput        bool
	}
}

type Module struct {
	ModuleName string
	Tasks      []GenTask
}

type CheckStatusResult struct {
	Status   string
	Expected string
	Current  string
}

type OutputResult struct {
	ResultType     string
	ResultContents string
	IsReady        bool
}

type RequestSaveWidgetStatus struct {
	ID      string `json:"id"`
	Payload string `json:"payload"`
}
