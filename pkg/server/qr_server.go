package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	qr "github.com/skip2/go-qrcode"
)

type QRServer struct {
	server *http.Server
	port   int
	done   chan struct{}
}

func NewQRServer() *QRServer {
	return &QRServer{
		done: make(chan struct{}),
	}
}

func (s *QRServer) Start(secret, accountName, issuer string) (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("failed to find available port: %w", err)
	}
	s.port = listener.Addr().(*net.TCPAddr).Port

	otpauthURL := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s", issuer, accountName, secret, issuer)

	qrPNG, err := qr.Encode(otpauthURL, qr.Medium, 256)
	if err != nil {
		listener.Close()
		return "", fmt.Errorf("failed to generate QR code: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>TOTP Setup - OpenPasswd</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            max-width: 500px;
            width: 100%%;
            padding: 40px;
            text-align: center;
        }
        h1 {
            color: #333;
            font-size: 28px;
            margin-bottom: 10px;
        }
        .subtitle {
            color: #666;
            font-size: 14px;
            margin-bottom: 30px;
        }
        .qr-container {
            background: #f8f9fa;
            border-radius: 12px;
            padding: 20px;
            margin-bottom: 30px;
            display: inline-block;
        }
        .qr-container img {
            display: block;
            width: 256px;
            height: 256px;
        }
        .manual-entry {
            background: #f8f9fa;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 20px;
            text-align: left;
        }
        .manual-entry h3 {
            color: #333;
            font-size: 14px;
            margin-bottom: 15px;
            text-align: center;
        }
        .field {
            margin-bottom: 12px;
        }
        .field-label {
            color: #666;
            font-size: 12px;
            font-weight: 600;
            text-transform: uppercase;
            margin-bottom: 4px;
        }
        .field-value {
            color: #333;
            font-size: 14px;
            font-family: 'Monaco', 'Courier New', monospace;
            background: white;
            padding: 8px 12px;
            border-radius: 4px;
            word-break: break-all;
            cursor: pointer;
            transition: background 0.2s;
        }
        .field-value:hover {
            background: #e9ecef;
        }
        .instructions {
            color: #666;
            font-size: 14px;
            line-height: 1.6;
            margin-bottom: 20px;
        }
        .step {
            margin-bottom: 10px;
        }
        .next-step {
            background: #667eea;
            color: white;
            font-size: 14px;
            padding: 12px 24px;
            border-radius: 8px;
            font-weight: 600;
            display: inline-block;
        }
        @media (max-width: 600px) {
            .container {
                padding: 30px 20px;
            }
            h1 {
                font-size: 24px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔐 TOTP Setup</h1>
        <p class="subtitle">Scan the QR code with your authenticator app</p>
        
        <div class="qr-container">
            <img src="/qr.png" alt="TOTP QR Code">
        </div>
        
        <div class="manual-entry">
            <h3>Or enter manually:</h3>
            <div class="field">
                <div class="field-label">Secret Key</div>
                <div class="field-value" onclick="copyToClipboard(this)" title="Click to copy">%s</div>
            </div>
            <div class="field">
                <div class="field-label">Account</div>
                <div class="field-value">%s</div>
            </div>
            <div class="field">
                <div class="field-label">Issuer</div>
                <div class="field-value">%s</div>
            </div>
        </div>
        
        <div class="instructions">
            <div class="step">1️⃣ Scan the QR code with Google Authenticator, Authy, or similar app</div>
            <div class="step">2️⃣ Return to your terminal</div>
            <div class="step">3️⃣ Enter the 6-digit code from your app</div>
        </div>
        
        <div class="next-step">✓ Return to terminal to continue</div>
    </div>
    
    <script>
        function copyToClipboard(element) {
            const text = element.textContent;
            navigator.clipboard.writeText(text).then(() => {
                const original = element.textContent;
                element.textContent = '✓ Copied!';
                setTimeout(() => {
                    element.textContent = original;
                }, 1500);
            });
        }
    </script>
</body>
</html>`, secret, accountName, issuer)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	})

	mux.HandleFunc("/qr.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(qrPNG)
	})

	s.server = &http.Server{
		Handler: mux,
	}

	go func() {
		_ = s.server.Serve(listener)
	}()

	url := fmt.Sprintf("http://127.0.0.1:%d", s.port)
	return url, nil
}

func (s *QRServer) Stop() error {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}
	return nil
}

func OpenBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		for _, browser := range []string{"xdg-open", "sensible-browser", "firefox", "chromium", "google-chrome"} {
			if _, err := exec.LookPath(browser); err == nil {
				cmd = exec.Command(browser, url)
				break
			}
		}
		if cmd == nil {
			return fmt.Errorf("no browser found")
		}
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
