package pongo2

/* Missing filters:

   escapejs
   force_escape
   iriencode
   linebreaks
   linenumbers
   phone2numeric
   safeseq
   slice
   truncatechars_html
   truncatewords
   truncatewords_html
   unordered_list
   urlize
   urlizetrunc
   wordwrap

   Filters that won't be added:

   get_static_prefix (reason: web-framework specific)
   pprint (reason: python-specific)
   static (reason: web-framework specific)

   Rethink:

   dictsort (python-specific; maybe one could add a filter to sort a list of structs by a specific field name)
   dictsortreversed (see dictsort)

   Filters that are provided through github.com/flosch/pongo2-addons:

   filesizeformat
   slugify
   timesince
   timeuntil
*/

import (
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())

	RegisterFilter("escape", filterEscape)
	RegisterFilter("safe", filterSafe)

	RegisterFilter("add", filterAdd)
	RegisterFilter("addslashes", filterAddslashes)
	RegisterFilter("capfirst", filterCapfirst)
	RegisterFilter("center", filterCenter)
	RegisterFilter("cut", filterCut)
	RegisterFilter("date", filterDate)
	RegisterFilter("default", filterDefault)
	RegisterFilter("default_if_none", filterDefaultIfNone)
	RegisterFilter("divisibleby", filterDivisibleby)
	RegisterFilter("first", filterFirst)
	RegisterFilter("floatformat", filterFloatformat)
	RegisterFilter("get_digit", filterGetdigit)
	RegisterFilter("join", filterJoin)
	RegisterFilter("last", filterLast)
	RegisterFilter("length", filterLength)
	RegisterFilter("length_is", filterLengthis)
	RegisterFilter("linebreaksbr", filterLinebreaksbr)
	RegisterFilter("ljust", filterLjust)
	RegisterFilter("lower", filterLower)
	RegisterFilter("make_list", filterMakelist)
	RegisterFilter("pluralize", filterPluralize)
	RegisterFilter("random", filterRandom)
	RegisterFilter("removetags", filterRemovetags)
	RegisterFilter("rjust", filterRjust)
	RegisterFilter("title", filterTitle)
	RegisterFilter("upper", filterUpper)
	RegisterFilter("urlencode", filterUrlencode)
	RegisterFilter("stringformat", filterStringformat)
	RegisterFilter("striptags", filterStriptags)
	RegisterFilter("time", filterDate) // time uses filterDate (same golang-format)
	RegisterFilter("truncatechars", filterTruncatechars)
	RegisterFilter("wordcount", filterWordcount)
	RegisterFilter("yesno", filterYesno)

	RegisterFilter("float", filterFloat)     // pongo-specific
	RegisterFilter("integer", filterInteger) // pongo-specific
}

func filterTruncatechars(in *Value, param *Value) (*Value, error) {
	s := in.String()
	newLen := param.Integer()
	if newLen < len(s) {
		if newLen >= 3 {
			return AsValue(fmt.Sprintf("%s...", s[:newLen-3])), nil
		}
		// Not enough space for the ellipsis
		return AsValue(s[:newLen]), nil
	}
	return in, nil
}

func filterEscape(in *Value, param *Value) (*Value, error) {
	output := strings.Replace(in.String(), "&", "&amp;", -1)
	output = strings.Replace(output, ">", "&gt;", -1)
	output = strings.Replace(output, "<", "&lt;", -1)
	output = strings.Replace(output, "\"", "&quot;", -1)
	output = strings.Replace(output, "'", "&#39;", -1)
	return AsValue(output), nil
}

func filterSafe(in *Value, param *Value) (*Value, error) {
	return in, nil // nothing to do here, just to keep track of the safe application
}

func filterAdd(in *Value, param *Value) (*Value, error) {
	if in.IsNumber() && param.IsNumber() {
		if in.IsFloat() || param.IsFloat() {
			return AsValue(in.Float() + param.Float()), nil
		} else {
			return AsValue(in.Integer() + param.Integer()), nil
		}
	}
	// If in/param is not a number, we're relying on the
	// Value's String() convertion and just add them both together
	return AsValue(in.String() + param.String()), nil
}

func filterAddslashes(in *Value, param *Value) (*Value, error) {
	output := strings.Replace(in.String(), "\\", "\\\\", -1)
	output = strings.Replace(output, "\"", "\\\"", -1)
	output = strings.Replace(output, "'", "\\'", -1)
	return AsValue(output), nil
}

func filterCut(in *Value, param *Value) (*Value, error) {
	return AsValue(strings.Replace(in.String(), param.String(), "", -1)), nil
}

func filterLength(in *Value, param *Value) (*Value, error) {
	return AsValue(in.Len()), nil
}

func filterLengthis(in *Value, param *Value) (*Value, error) {
	return AsValue(in.Len() == param.Integer()), nil
}

func filterDefault(in *Value, param *Value) (*Value, error) {
	if !in.IsTrue() {
		return param, nil
	}
	return in, nil
}

func filterDefaultIfNone(in *Value, param *Value) (*Value, error) {
	if in.IsNil() {
		return param, nil
	}
	return in, nil
}

func filterDivisibleby(in *Value, param *Value) (*Value, error) {
	if param.Integer() == 0 {
		return AsValue(false), nil
	}
	return AsValue(in.Integer()%param.Integer() == 0), nil
}

func filterFirst(in *Value, param *Value) (*Value, error) {
	if in.CanSlice() && in.Len() > 0 {
		return in.Index(0), nil
	}
	return AsValue(""), nil
}

func filterFloatformat(in *Value, param *Value) (*Value, error) {
	val := in.Float()

	decimals := -1
	if !param.IsNil() {
		// Any argument provided?
		decimals = param.Integer()
	}

	// if the argument is not a number (e. g. empty), the default
	// behaviour is trim the result
	trim := !param.IsNumber()

	if decimals <= 0 {
		// argument is negative or zero, so we
		// want the output being trimmed
		decimals = -decimals
		trim = true
	}

	if trim {
		// Remove zeroes
		if float64(int(val)) == val {
			return AsValue(in.Integer()), nil
		}
	}

	return AsValue(strconv.FormatFloat(val, 'f', decimals, 64)), nil
}

func filterGetdigit(in *Value, param *Value) (*Value, error) {
	i := param.Integer()
	l := len(in.String()) // do NOT use in.Len() here!
	if i <= 0 || i > l {
		return in, nil
	}
	return AsValue(in.String()[l-i] - 48), nil
}

func filterJoin(in *Value, param *Value) (*Value, error) {
	if !in.CanSlice() {
		return in, nil
	}
	sep := param.String()
	sl := make([]string, 0, in.Len())
	for i := 0; i < in.Len(); i++ {
		sl = append(sl, in.Index(i).String())
	}
	return AsValue(strings.Join(sl, sep)), nil
}

func filterLast(in *Value, param *Value) (*Value, error) {
	if in.CanSlice() && in.Len() > 0 {
		return in.Index(in.Len() - 1), nil
	}
	return AsValue(""), nil
}

func filterUpper(in *Value, param *Value) (*Value, error) {
	return AsValue(strings.ToUpper(in.String())), nil
}

func filterLower(in *Value, param *Value) (*Value, error) {
	return AsValue(strings.ToLower(in.String())), nil
}

func filterMakelist(in *Value, param *Value) (*Value, error) {
	s := in.String()
	result := make([]string, 0, len(s))
	for _, c := range s {
		result = append(result, string(c))
	}
	return AsValue(result), nil
}

func filterCapfirst(in *Value, param *Value) (*Value, error) {
	if in.Len() <= 0 {
		return AsValue(""), nil
	}
	t := in.String()
	return AsValue(strings.ToUpper(string(t[0])) + t[1:]), nil
}

func filterCenter(in *Value, param *Value) (*Value, error) {
	width := param.Integer()
	slen := in.Len()
	if width <= slen {
		return in, nil
	}

	spaces := width - slen
	left := spaces/2 + spaces%2
	right := spaces / 2

	return AsValue(fmt.Sprintf("%s%s%s", strings.Repeat(" ", left),
		in.String(), strings.Repeat(" ", right))), nil
}

func filterDate(in *Value, param *Value) (*Value, error) {
	t, is_time := in.Interface().(time.Time)
	if !is_time {
		return nil, errors.New("Filter input argument must be of type 'time.Time'.")
	}
	return AsValue(t.Format(param.String())), nil
}

func filterFloat(in *Value, param *Value) (*Value, error) {
	return AsValue(in.Float()), nil
}

func filterInteger(in *Value, param *Value) (*Value, error) {
	return AsValue(in.Integer()), nil
}

func filterLinebreaksbr(in *Value, param *Value) (*Value, error) {
	return AsValue(strings.Replace(in.String(), "\n", "<br />", -1)), nil
}

func filterLjust(in *Value, param *Value) (*Value, error) {
	times := param.Integer() - in.Len()
	if times < 0 {
		times = 0
	}
	return AsValue(fmt.Sprintf("%s%s", in.String(), strings.Repeat(" ", times))), nil
}

func filterUrlencode(in *Value, param *Value) (*Value, error) {
	return AsValue(url.QueryEscape(in.String())), nil
}

func filterStringformat(in *Value, param *Value) (*Value, error) {
	return AsValue(fmt.Sprintf(param.String(), in.Interface())), nil
}

var re_striptags = regexp.MustCompile("<[^>]*?>")

func filterStriptags(in *Value, param *Value) (*Value, error) {
	s := in.String()

	// Strip all tags
	s = re_striptags.ReplaceAllString(s, "")

	return AsValue(strings.TrimSpace(s)), nil
}

func filterPluralize(in *Value, param *Value) (*Value, error) {
	if in.IsNumber() {
		// Works only on numbers
		if param.Len() > 0 {
			endings := strings.Split(param.String(), ",")
			if len(endings) > 2 {
				return nil, errors.New("You cannot pass more than 2 arguments to filter 'pluralize'.")
			}
			if len(endings) == 1 {
				// 1 argument
				if in.Integer() != 1 {
					return AsValue(endings[0]), nil
				}
			} else {
				if in.Integer() != 1 {
					// 2 arguments
					return AsValue(endings[1]), nil
				}
				return AsValue(endings[0]), nil
			}
		} else {
			if in.Integer() != 1 {
				// return default 's'
				return AsValue("s"), nil
			}
		}

		return AsValue(""), nil
	} else {
		return nil, errors.New("Filter 'pluralize' does only work on numbers.")
	}
}

func filterRandom(in *Value, param *Value) (*Value, error) {
	if !in.CanSlice() || in.Len() <= 0 {
		return in, nil
	}
	i := rand.Intn(in.Len())
	return in.Index(i), nil
}

func filterRemovetags(in *Value, param *Value) (*Value, error) {
	s := in.String()
	tags := strings.Split(param.String(), ",")

	// Strip only specific tags
	for _, tag := range tags {
		re := regexp.MustCompile(fmt.Sprintf("</?%s/?>", tag))
		s = re.ReplaceAllString(s, "")
	}

	return AsValue(strings.TrimSpace(s)), nil
}

func filterRjust(in *Value, param *Value) (*Value, error) {
	return AsValue(fmt.Sprintf(fmt.Sprintf("%%%ds", param.Integer()), in.String())), nil
}

func filterTitle(in *Value, param *Value) (*Value, error) {
	if !in.IsString() {
		return AsValue(""), nil
	}
	return AsValue(strings.Title(strings.ToLower(in.String()))), nil
}

func filterWordcount(in *Value, param *Value) (*Value, error) {
	return AsValue(len(strings.Fields(in.String()))), nil
}

func filterYesno(in *Value, param *Value) (*Value, error) {
	choices := map[int]string{
		0: "yes",
		1: "no",
		2: "maybe",
	}
	param_string := param.String()
	custom_choices := strings.Split(param_string, ",")
	if len(param_string) > 0 {
		if len(custom_choices) > 3 {
			return nil, errors.New(fmt.Sprintf("You cannot pass more than 3 options to the 'yesno'-filter (got: '%s').", param_string))
		}
		if len(custom_choices) < 2 {
			return nil, errors.New(fmt.Sprintf("You must pass either no or at least 2 arguments to the 'yesno'-filter (got: '%s').", param_string))
		}

		// Map to the options now
		choices[0] = custom_choices[0]
		choices[1] = custom_choices[1]
		if len(custom_choices) == 3 {
			choices[2] = custom_choices[2]
		}
	}

	// maybe
	if in.IsNil() {
		return AsValue(choices[2]), nil
	}

	// yes
	if in.IsTrue() {
		return AsValue(choices[0]), nil
	}

	// no
	return AsValue(choices[1]), nil
}
