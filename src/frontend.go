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
            /* Dracula Theme Colors */
            --bg-color: #282a36;
            --fg-color: #f8f8f2;
            --comment-color: #6272a4;
            --cyan: #8be9fd;
            --green: #50fa7b;
            --pink: #ff79c6;
            --purple: #bd93f9;
            --font-family: 'Courier New', Courier, monospace;
        }
        body {
            background-color: var(--bg-color);
            color: var(--fg-color);
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
            color: var(--purple);
            border-bottom: 2px solid var(--purple);
            padding-bottom: 0.5rem;
            margin-bottom: 0.5rem;
        }
        p {
            font-size: 1.1rem;
        }
        .prompt::before {
            content: "menshen@terminus.re:~$ ";
            color: var(--green);
        }
        .blinking-cursor {
            display: inline-block;
            width: 10px;
            height: 1.2em;
            background-color: var(--fg-color);
            vertical-align: middle;
            animation: blink 1s step-end infinite;
        }
        @keyframes blink {
            0%, 100% { opacity: 1; }
            50% { opacity: 0; }
        }
        
        .interactive-prompt {
            margin-top: 2rem;
            display: inline-block;
            cursor: pointer;
            padding: 0.5rem;
            border-radius: 4px;
            transition: background-color 0.2s;
            user-select: none;
            outline: none;
        }
        .interactive-prompt:focus {
            background-color: #44475a;
            box-shadow: 0 0 0 2px var(--purple);
        }
        .interactive-prompt:hover {
            background-color: #44475a;
        }
        .interactive-prompt:active {
            background-color: #6272a4;
        }
        .interactive-prompt.executing .command {
            color: var(--pink);
        }
        
        pre {
            background-color: #44475a;
            padding: 1rem;
            border-left: 3px solid var(--pink);
            overflow-x: auto;
            color: var(--fg-color);
            border-radius: 4px;
        }
        a {
            color: var(--pink);
            text-decoration: none;
            border-bottom: 1px dashed var(--pink);
        }
        a:hover {
            color: var(--cyan);
            border-bottom-color: var(--cyan);
        }
        .hidden {
            display: none;
        }
        .comment {
            color: var(--comment-color);
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Menshen</h1>
        <p>Ephemeral, anonymous git hosting. Repositories self-destruct after inactivity.</p>
        <p style="font-size: 0.9rem;">[ <a href="https://github.com/saalik/menshen" target="_blank">Source Code</a> ]</p>
        
        <div id="createBtn" class="interactive-prompt" tabindex="0" role="button" title="Click or press Enter to run command">
            <span class="prompt"></span><span class="command">/new</span><span class="blinking-cursor"></span>
        </div>
        
        <div id="result" class="hidden" style="margin-top: 2rem;">
            <p class="comment">> Repository created successfully. Clone URL:</p>
            <pre id="repoUrl"></pre>
            <p class="comment">> Quickstart:</p>
            <pre id="repoUsage"></pre>
        </div>
    </div>

    <script>
        document.getElementById('createBtn').addEventListener('click', async () => {
            const btn = document.getElementById('createBtn');
            if (btn.classList.contains('executing')) return; // Prevent multiple clicks
            
            btn.classList.add('executing');
            const commandSpan = btn.querySelector('.command');
            commandSpan.innerText = "/new --executing...";
            
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
                commandSpan.innerText = "/new --done";
                btn.querySelector('.blinking-cursor').style.display = 'none';
                btn.style.cursor = 'default';
                btn.style.pointerEvents = 'none'; // Disable hover and click after completion
                btn.blur();
            } catch (err) {
                alert('Error creating repository');
                btn.classList.remove('executing');
                commandSpan.innerText = "/new";
            }
        });
        
        // Add Enter key support for accessibility
        document.getElementById('createBtn').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                e.preventDefault();
                this.click();
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
