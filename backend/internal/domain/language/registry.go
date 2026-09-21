package language

// Language code type following ISO-639-1/3 or BCP-47 conventions
type Code string

const (
	English  Code = "en"
	Hindi    Code = "hi"
	Bengali  Code = "bn"
	Telugu   Code = "te"
	Marathi  Code = "mr"
	Tamil    Code = "ta"
	Urdu     Code = "ur"
	Gujarati Code = "gu"
	Kannada  Code = "kn"
	Odia     Code = "or"
	Malayalam Code = "ml"
	Punjabi  Code = "pa"
	Assamese Code = "as"
	Maithili Code = "mai"
	Santali  Code = "sat"
	Kashmiri Code = "ks"
	Nepali   Code = "ne"
	Sindhi   Code = "sd"
	Konkani  Code = "kok"
	Dogri    Code = "doi"
	Manipuri Code = "mni"
	Bodo     Code = "brx"
	Sanskrit Code = "sa"
)

var ScheduledLanguages = []Code{
	Hindi, Bengali, Telugu, Marathi, Tamil, Urdu, Gujarati, Kannada, Odia,
	Malayalam, Punjabi, Assamese, Maithili, Santali, Kashmiri, Nepali,
	Sindhi, Konkani, Dogri, Manipuri, Bodo, Sanskrit, English,
}
