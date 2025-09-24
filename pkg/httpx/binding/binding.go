package binding

import (
    "encoding/json"
    "errors"
    "net/http"
    "strings"

    "github.com/go-playground/validator/v10"
    "github.com/gorilla/schema"
)

var (
    validate     = validator.New(validator.WithRequiredStructEnabled())
    queryDecoder = func() *schema.Decoder {
        d := schema.NewDecoder()
        d.IgnoreUnknownKeys(true)
        d.SetAliasTag("json")
        return d
    }()
)

// BindJSON decodes a JSON body into dst.
func BindJSON(r *http.Request, dst any) error {
    if r.Body == nil {
        return errors.New("empty body")
    }
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    return dec.Decode(dst)
}

// BindQuery decodes URL query parameters into dst.
func BindQuery(r *http.Request, dst any) error {
    return queryDecoder.Decode(dst, r.URL.Query())
}

// Validate runs struct validation and returns a map of field->message on failure.
func Validate(v any) (map[string]string, error) {
    if err := validate.Struct(v); err != nil {
        verrs, ok := err.(validator.ValidationErrors)
        if !ok {
            return nil, err
        }
        out := make(map[string]string, len(verrs))
        for _, fe := range verrs {
            field := jsonFieldName(fe)
            out[field] = messageFor(fe)
        }
        return out, nil
    }
    return nil, nil
}

func jsonFieldName(fe validator.FieldError) string {
    if tag := fe.StructField(); tag != "" {
        // prefer json tag if available
        if fe.StructField() != "" {
            if jsonTag := fe.Field(); jsonTag != "" {
                // validator doesn't expose json tag directly; fallback to lowercase field
                return strings.ToLower(fe.Field())
            }
        }
        return strings.ToLower(tag)
    }
    return fe.Field()
}

func messageFor(fe validator.FieldError) string {
    switch fe.Tag() {
    case "required":
        return "is required"
    case "email":
        return "must be a valid email"
    case "min":
        return "must be at least " + fe.Param()
    case "max":
        return "must be at most " + fe.Param()
    case "len":
        return "must be length " + fe.Param()
    case "oneof":
        return "must be one of: " + fe.Param()
    case "url":
        return "must be a valid URL"
    case "uuid":
        return "must be a valid UUID"
    case "numeric":
        return "must be numeric"
    case "alphanum":
        return "must be alphanumeric"
    case "gt":
        return "must be greater than " + fe.Param()
    case "gte":
        return "must be greater or equal to " + fe.Param()
    case "lt":
        return "must be less than " + fe.Param()
    case "lte":
        return "must be less or equal to " + fe.Param()
    default:
        return fe.Tag()
    }
}

