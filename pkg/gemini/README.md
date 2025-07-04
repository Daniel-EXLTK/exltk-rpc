# Gemini Client

Cliente HTTP para interactuar con la API de Gemini de Google. Esta biblioteca proporciona una interfaz simple para generar contenido usando el LLM Gemini.

## Características

- **HTTP Client** nativo sin dependencias externas
- **Autenticación** con API Key
- **Timeout** configurable
- **Retry Logic** para manejo de errores
- **Error Handling** robusto
- **Logging** estructurado

## Uso

```go
import "github.com/Daniel-EXLTK/exltk-rpc/pkg/gemini"

// Crear cliente
client := gemini.NewClient("your-api-key")

// Generar contenido
response, err := client.Generate("¿Cuál es la capital de Francia?")
if err != nil {
    log.Fatal(err)
}

fmt.Println(response)
```

## Configuración

### Variables de Entorno

```bash
export GEMINI_API_KEY="your-api-key-here"
export GEMINI_TIMEOUT="30s"
export GEMINI_MAX_RETRIES="3"
```

### Opciones de Cliente

```go
client := gemini.NewClient("api-key",
    gemini.WithTimeout(30*time.Second),
    gemini.WithMaxRetries(3),
    gemini.WithBaseURL("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-exp:generateContent"),
)
```

## Arquitectura

```
pkg/gemini/
├── client.go           # Cliente HTTP principal
├── types.go            # Tipos Request/Response
├── config.go           # Configuración del cliente
├── client_test.go      # Tests unitarios
└── README.md           # Documentación
```

## API Reference

### Client

```go
type Client struct {
    apiKey     string
    httpClient *http.Client
    baseURL    string
    timeout    time.Duration
    maxRetries int
}
```

### Methods

```go
func NewClient(apiKey string, options ...Option) *Client
func (c *Client) Generate(prompt string) (string, error)
func (c *Client) GenerateWithContext(ctx context.Context, prompt string) (string, error)
```

## Error Handling

```go
if err != nil {
    switch {
    case errors.Is(err, gemini.ErrAPIKeyMissing):
        log.Fatal("API key is required")
    case errors.Is(err, gemini.ErrRateLimited):
        log.Println("Rate limited, retrying...")
    case errors.Is(err, gemini.ErrQuotaExceeded):
        log.Fatal("Quota exceeded")
    default:
        log.Printf("API error: %v", err)
    }
}
```

## Testing

```go
// Mock client para testing
mockClient := gemini.NewMockClient()
mockClient.SetResponse("Mocked response")

response, err := mockClient.Generate("test prompt")
// response == "Mocked response"
```

## Limitaciones

- **Rate Limits**: Respeta los límites de la API de Gemini
- **Context Length**: Máximo de tokens por request
- **API Keys**: Requiere API key válida de Google AI 