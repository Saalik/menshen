package main

import (
	"net/http"
)

const frontendHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Menshen - Ephemeral Git</title>
    <style>
        :root {
            --bg-color: #0d0d0d;
            --text-color: #00ff00;
            --font-family: 'Courier New', Courier, monospace;
        }
        body {
            background-color: var(--bg-color);
            color: var(--text-color);
            font-family: var(--font-family);
            padding: 2rem;
            line-height: 1.6;
            margin: 0;
            display: flex;
            justify-content: center;
        }
        .container {
            max-width: 800px;
            width: 100%;
        }
        h1 {
            border-bottom: 2px solid var(--text-color);
            padding-bottom: 0.5rem;
            text-shadow: 0 0 5px var(--text-color);
        }
        p {
            font-size: 1.1rem;
        }
        .prompt::before {
            content: "guest@menshen:~$ ";
            color: #aaaaaa;
        }
        .blinking-cursor {
            display: inline-block;
            width: 10px;
            height: 1.2em;
            background-color: var(--text-color);
            vertical-align: middle;
            animation: blink 1s step-end infinite;
            box-shadow: 0 0 5px var(--text-color);
        }
        @keyframes blink {
            0%, 100% { opacity: 1; }
            50% { opacity: 0; }
        }
        .btn {
            background-color: transparent;
            color: var(--text-color);
            border: 1px solid var(--text-color);
            padding: 0.5rem 1rem;
            font-family: inherit;
            cursor: pointer;
            margin-top: 1.5rem;
            font-size: 1rem;
            transition: all 0.2s;
            box-shadow: 0 0 5px rgba(0, 255, 0, 0.2);
        }
        .btn:hover {
            background-color: var(--text-color);
            color: var(--bg-color);
            box-shadow: 0 0 10px var(--text-color);
        }
        pre {
            background-color: #1a1a1a;
            padding: 1rem;
            border-left: 3px solid var(--text-color);
            overflow-x: auto;
            color: #00cc00;
        }
        .hidden {
            display: none;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Menshen</h1>
        <p>Ephemeral, anonymous git hosting. Repositories self-destruct after inactivity.</p>
        <p style="font-size: 0.9rem; opacity: 0.8;">[ Source Code: <a href="https://github.com/saalik/menshen" target="_blank" style="color: var(--text-color); text-decoration: none; border-bottom: 1px dashed var(--text-color);">github.com/saalik/menshen</a> ]</p>
        
        <div style="margin-top: 2rem;">
            <span class="prompt"></span><span>./create_repo.sh</span><span class="blinking-cursor"></span>
        </div>
        
        <button id="createBtn" class="btn">[ Create Repository ]</button>
        
        <div id="result" class="hidden" style="margin-top: 2rem;">
            <p>> Repository created successfully. Clone URL:</p>
            <pre id="repoUrl"></pre>
            <p>> Quickstart:</p>
            <pre id="repoUsage"></pre>
        </div>
    </div>

    <script>
        document.getElementById('createBtn').addEventListener('click', async () => {
            const btn = document.getElementById('createBtn');
            btn.disabled = true;
            btn.innerText = "[ Creating... ]";
            
            try {
                const response = await fetch('/new', { method: 'POST' });
                if (!response.ok) throw new Error('Failed to create');
                const url = await response.text();
                const cleanUrl = url.trim();
                
                document.getElementById('repoUrl').innerText = cleanUrl;
                document.getElementById('repoUsage').innerText = 
"git clone " + cleanUrl + " repo\n" +
"cd repo\n" +
"touch README.md\n" +
"git add README.md\n" +
"git commit -m \"Initial commit\"\n" +
"git push origin master";
                
                document.getElementById('result').classList.remove('hidden');
                btn.style.display = 'none';
                document.querySelector('.blinking-cursor').style.display = 'none';
            } catch (err) {
                alert('Error creating repository');
                btn.disabled = false;
                btn.innerText = "[ Create Repository ]";
            }
        });
    </script>
</body>
</html>`

func (s *Server) handleFrontend(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(frontendHTML))
}
