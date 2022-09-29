package gtype

const (
	ContentTypeHtml     = "text/html"
	ContentTypeText     = "text/plain"
	ContentTypeJson     = "application/json"
	ContentTypeFormData = "multipart/form-data"
	ContentTypeXml      = "application/xml"
	ContentTypeSoap     = "application/soap+xml"
)
const (
	FormValueKindText = 0
	FormValueKindFile = 1
)

type Appendix interface {
	Add(value interface{}, name, note string, example interface{})
	AddItem(item AppendixItem)
}

type AppendixItem interface {
	Value() interface{}
	Name() string
	Note() string
	Example() interface{}
}

type Function interface {
	SetNote(v string)
	SetRemark(v string)
	AddInputHeader(required bool, name, note, defaultValue string, optionValues ...string)
	ClearInputHeader()
	AddInputQuery(required bool, name, note, defaultValue string, optionValues ...string)
	RemoveInputQuery(name string)
	AddInputForm(required bool, key, note string, valueKind int, defaultValue interface{})
	RemoveInputForm(key string)
	SetInputFormat(v int)
	SetInputExample(v interface{})
	SetInputJsonExample(v interface{})
	SetInputXmlExample(v interface{})
	SetCustomInputJsonExample(v interface{})
	SetInputAppendix(label string) Appendix
	AddOutputHeader(name, value string)
	ClearOutputHeader()
	SetOutputFormat(v int)
	SetOutputExample(v interface{})
	SetOutputDataExample(v interface{})
	SetCustomOutputDataExample(v interface{})
	SetCustomOutputDatabaseExample(v interface{})
	SetOutputXmlExample(v interface{})
	AddOutputError(err Error)
	AddOutputErrorCustom(code int, summary string)
	SetOutputAppendix(label string) Appendix
}

type Catalog interface {
	AddChild(name string) Catalog
	AddFunction(method string, uri Uri, name string) Function
}

type Doc interface {
	Enable() bool
	AddCatalog(name string) Catalog
	Catalogs() interface{}
	Function(id, schema, host string) (interface{}, error)
	OnFunctionReady(f func(index int, method, path, name string))
	TokenUI(id string) (interface{}, error)
	TokenCreate(id string, items []TokenAuth, ctx Context) (string, Error)
	Log(handle DocHandle, method string, uri Uri)
	Regenerate()
}
