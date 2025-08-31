# Mermaid Agent Documenter

An intelligent CLI tool that uses AI agents to generate comprehensive Mermaid diagrams and documentation from application transcripts, walkthroughs, and recordings.

## ✨ Features

- 🤖 **AI-Powered Analysis** - Uses advanced LLMs to understand application workflows from transcripts
- 📊 **Multiple Diagram Types** - Generates sequence, flowchart, class, ER, state, and journey diagrams
- 🔧 **Multi-Provider Support** - Works with OpenAI, Anthropic Claude, and Google Gemini
- 🛡️ **Safe & Configurable** - Built-in safety controls, confidence thresholds, and PII redaction
- 📁 **Structured Output** - Produces organized Markdown files with embedded Mermaid diagrams
- ⚡ **CLI-First Design** - Simple command-line interface with dry-run capabilities
- 🔄 **Tool Integration** - Extensible tool system for file operations and web fetching

## 📋 Requirements

- **Go 1.24+** - Required for building and running
- **API Key** - One of the following:
  - `OPENAI_API_KEY` for GPT models
  - `ANTHROPIC_API_KEY` for Claude models
  - `GOOGLE_API_KEY` for Gemini models

## 🚀 Installation

### Option 1: Download Pre-built Binary
```bash
# Download from GitHub releases (coming soon)
# curl -L https://github.com/landanqrew/mermaid-agent-documenter/releases/download/v1.0.0/mad-linux-amd64 -o mad
# chmod +x mad
```

### Option 2: Build from Source
```bash
# Clone the repository
git clone https://github.com/landanqrew/mermaid-agent-documenter.git
cd mermaid-agent-documenter

# Build the CLI tool
go build -o mad .

# (Optional) Move to PATH
sudo mv mad /usr/local/bin/
```

## ⚙️ Setup and Configuration

### 1. Initialize the Global Environment
```bash
# Initialize global configuration
./mad init
```

This creates:
- `~/mermaid-agent-documenter/config.json` - Global configuration file
- `~/mermaid-agent-documenter/logs/` - Global execution logs

### 2. Create a Project
```bash
# Create a new project in the current directory
./mad init my-project

# Or create a project for an e-commerce application
./mad init ecommerce-app
```

This creates:
- `my-project/` - Project directory
- `my-project/config.json` - Project reference in global config
- `my-project/transcripts/` - Place your transcript files here
- `my-project/out/` - Generated diagrams and documentation
- `my-project/logs/` - Project-specific execution logs

### 3. Set API Key
Choose one provider and set the corresponding environment variable:

```bash
# For OpenAI (default)
export OPENAI_API_KEY="your-openai-api-key-here"

# For Anthropic Claude
export ANTHROPIC_API_KEY="your-anthropic-api-key-here"

# For Google Gemini
export GOOGLE_API_KEY="your-google-api-key-here"
```

### 4. Configure Provider (Optional)
Edit `~/mermaid-agent-documenter/config.json` to change:
- Default provider (`"provider": "openai"`)
- Model selection per provider
- Safety settings
- Confidence thresholds

## 📖 Usage

### Basic Workflow

1. **Create a project** - Initialize a project for your application
2. **Add transcripts** - Place your application walkthrough files in the project's `transcripts/` directory
3. **Run the agent** - Use the CLI to analyze transcripts and generate documentation
4. **Review output** - Check the generated Markdown files with Mermaid diagrams in the `out/` directory

### Example Workflow

```bash
# 1. Create a project
./mad init my-auth-app

# 2. Add your transcript file
echo "This is a walkthrough of our user authentication system..." > my-auth-app/transcripts/auth-walkthrough.txt

# 3. Generate documentation
cd my-auth-app
../mad run auth-walkthrough.txt --dry-run  # Preview what will be generated

# 4. Generate the actual documentation
../mad run auth-walkthrough.txt

# 5. Check the output
ls out/docs/diagrams/
```

### Example Transcript File
Create `my-auth-app/transcripts/auth-walkthrough.txt`:

```
This is a walkthrough of our user authentication system.

When a user logs in, they provide their username and password to the login form.
The frontend sends a POST request to /api/login with the credentials.
The API gateway receives this request and forwards it to the authentication service.

The authentication service validates the credentials against the user database.
If the credentials are valid, it generates a JWT token and returns it.
The API gateway then sends the JWT back to the frontend, and the user is logged in.
```

### Generate Documentation
```bash
# From within your project directory
./mad run auth-walkthrough.txt --dry-run   # Preview generation
./mad run auth-walkthrough.txt             # Generate with confirmation
./mad run auth-walkthrough.txt --yes       # Skip confirmation
```

## 📋 Command Reference

### `mad init [project-name]`
Initialize a new project or the global environment.

```bash
./mad init                    # Initialize global environment
./mad init my-project         # Create new project called "my-project"
```

### `mad run [transcript]`
Run the agent on a transcript to generate documentation.

```bash
./mad run transcript.txt [flags]

Flags:
  --dry-run   Print planned actions without executing
  --yes       Skip confirmation prompts

Notes:
- If run from within a project directory, uses project's transcripts/ and out/ directories
- If no current project is set, uses global configuration
```

### `mad plan [transcript]`
Plan the agent's actions without executing (shows what would be generated).

```bash
./mad plan transcript.txt [flags]

Flags:
  --yes       Skip confirmation prompts
```

### `mad validate [path]`
Validate a generated manifest or Mermaid file for syntax correctness.

```bash
./mad validate docs/diagrams/auth/sequence-login.md
```

### `mad config secrets set <provider> <api-key>`
Set API key for a model provider.

```bash
./mad config secrets set openai "sk-your-openai-key-here"
./mad config secrets set anthropic "sk-ant-your-anthropic-key"
./mad config secrets set google "your-google-api-key"
```

### `mad config secrets list`
List configured API keys (without showing actual keys).

```bash
./mad config secrets list
```

### `mad config provider set <provider>`
Set the default LLM provider.

```bash
./mad config provider set openai      # Use OpenAI models
./mad config provider set anthropic   # Use Anthropic models
./mad config provider set google      # Use Google models
```

### `mad config provider list`
List available providers and current selection.

```bash
./mad config provider list
```

### `mad config model set <model>`
Set the model for the current provider.

```bash
# For OpenAI
./mad config model set gpt-4o
./mad config model set gpt-4o-mini

# For Anthropic
./mad config model set claude-3-5-sonnet-20241022
./mad config model set claude-3-haiku-20240307

# For Google
./mad config model set gemini-1.5-pro
./mad config model set gemini-1.5-flash

# You can also use any custom model name
./mad config model set my-custom-model
```

**Note**: You can use any model name that the provider supports. The system will attempt to use it even if it's not in our known models list.

### `mad config model list`
List available models for the current provider.

```bash
./mad config model list
```

Shows both "known" models (from our curated list) and "custom" models you've configured.

### `mad config model refresh`
Query provider APIs for current model availability.

```bash
./mad config model refresh
```

**Features**:
- Connects to provider APIs to get live model lists
- Falls back to known models if API is unavailable
- Shows new models not in our curated list
- Displays model status (known, custom, new)
- No API key required (uses known models as fallback)

**Example Output**:
```
🔄 Refreshing models for Openai...
📡 Fetching from provider API...
✅ Found 15 models from API:

📋 Known Models (available via API):
✅ gpt-4o (current)
○ gpt-4o-mini
○ gpt-4-turbo

🆕 New/Discovered Models:
○ gpt-4o-new-variant
○ gpt-4-turbo-new

💡 Tip: You can use these new models with:
   mad config model set <model-name>
```

### `mad config project set <project-directory>`
Set the current project directory.

```bash
./mad config project set /path/to/my-project
./mad config project set ./my-auth-app
```

## 📁 Output Structure

### Project-Based Organization
When using projects, generated files are organized in your project's `out/` directory:

```
my-project/
├── transcripts/
│   └── auth-walkthrough.txt
├── out/
│   ├── docs/
│   │   └── diagrams/
│   │       ├── auth/
│   │       │   ├── sequence-login.md
│   │       │   └── flowchart-auth.md
│   │       ├── user/
│   │       │   └── class-user-model.md
│   │       └── api/
│   │           └── sequence-api-flow.md
│   └── logs/
│       └── execution-log.jsonl
└── config.json (references global config)
```

### Global Organization (Legacy)
If no project is used, files go to `~/mermaid-agent-documenter/output/`:

```
~/mermaid-agent-documenter/
├── output/
│   └── docs/
│       └── diagrams/
│           ├── auth/
│           │   ├── sequence-login.md
│           │   └── flowchart-auth.md
│           └── ...
└── config.json
```

### Example Generated File
```markdown
---
title: User Login and Token Issuance
area: auth
tags: [sequence, auth]
---

## Context
Sequence covering credential verification and JWT issuance via API Gateway and Auth service.

```mermaid
sequenceDiagram
  participant User
  participant API as API Gateway
  participant Auth
  User->>API: POST /login
  API->>Auth: Verify credentials
  Auth-->>API: JWT
  API-->>User: 200 OK (JWT)
```
```

## 🔧 Configuration Options

### Global Configuration
The main `~/mermaid-agent-documenter/config.json` file supports the following settings:

```json
{
  "provider": "openai",           // Default LLM provider
  "models": {                     // Model selection per provider
    "openai": "gpt-5-mini",
    "anthropic": "claude-3.5-sonnet",
    "google": "gemini-2.5-flash"
  },
  "log": {
    "level": "info",              // Logging level
    "redact": true,               // Redact sensitive data
    "storeChainOfThought": false  // Store AI reasoning (optional)
  },
  "safety": {
    "mode": "standard",           // Safety mode: strict|standard|off
    "piiRedaction": true          // Auto-redact PII
  },
  "limits": {
    "maxSteps": 12,               // Max agent steps per run
    "runTimeoutSec": 300,         // Timeout in seconds
    "tokenBudget": 100000,        // Max tokens per run
    "costCeilingUsd": 1.0         // Max cost per run
  },
  "confidenceThreshold": 0.90,    // Min confidence for file writes
  "outDir": "~/mermaid-agent-documenter/output",
  "currentProject": {             // Currently active project
    "name": "my-auth-app",
    "rootDir": "/path/to/my-auth-app",
    "createdAt": "2025-01-01T12:00:00Z"
  }
}
```

### Project Structure
When you create a project with `mad init my-project`, it creates:

- **Project Directory**: Contains all project-specific files
- **Transcripts Directory**: `my-project/transcripts/` - Place your transcript files here
- **Output Directory**: `my-project/out/` - Generated documentation goes here
- **Logs Directory**: `my-project/logs/` - Project-specific execution logs

### Switching Projects
Projects are managed through the global config. The `currentProject` field determines which project's directories are used for transcripts and output.

To switch projects, you would manually edit the global `config.json` or reinitialize with a different project name.

## 🎯 Supported Diagram Types

The agent automatically selects the most appropriate diagram type:

- **Sequence Diagrams** - For request/response flows, API calls, authentication
- **Flowcharts** - For business logic, decision trees, error handling
- **Class Diagrams** - For data models, object relationships
- **ER Diagrams** - For database schemas and relationships
- **State Diagrams** - For lifecycle management, state transitions
- **Journey Diagrams** - For user journeys and workflows
- **Graph Diagrams** - For deployment topology and connectivity

## 🚨 Troubleshooting

### Common Issues

**"API key not found"**
```bash
# Check if environment variable is set
echo $OPENAI_API_KEY

# Set the API key
export OPENAI_API_KEY="your-key-here"
```

**"Command not found"**
```bash
# Make sure the binary is executable and in PATH
chmod +x ./mad
./mad --help

# Or move to PATH
sudo cp mad /usr/local/bin/mad
```

**"Build fails"**
```bash
# Ensure Go 1.24+ is installed
go version

# Clean and rebuild
go clean
go mod tidy
go build -o mad .
```

**"Agent execution failed"**
- Check your API key is valid and has sufficient credits
- Try with a smaller transcript file first
- Use `--dry-run` to test without API calls
- Check the logs in `~/mermaid-agent-documenter/logs/`

### Getting Help
```bash
# Show all commands
./mad --help

# Show command-specific help
./mad run --help
./mad init --help
```

## 📊 Examples

### Simple Authentication Flow
**Input Transcript:**
```
Users log in by entering username/password. The app sends credentials to /auth endpoint, which validates against the database and returns a JWT token.
```

**Generated Output:**
- `docs/diagrams/auth/sequence-login.md` - Login sequence diagram
- `docs/diagrams/auth/flowchart-auth-validation.md` - Authentication validation flow

### Complex E-commerce Checkout
**Input Transcript:**
```
Checkout process starts when user clicks 'Buy Now'. Cart items are validated, payment is processed via Stripe, order is created in database, confirmation email is sent.
```

**Generated Output:**
- `docs/diagrams/checkout/sequence-checkout-flow.md` - Complete checkout sequence
- `docs/diagrams/payment/flowchart-payment-processing.md` - Payment validation logic
- `docs/diagrams/order/state-order-lifecycle.md` - Order state transitions

## 🧠 Model Management

The CLI includes smart model management that adapts to the rapidly changing AI landscape:

### **Flexible Model Support**
- **Known Models**: Curated list of popular, stable models for each provider
- **Custom Models**: Support for any model name the provider accepts
- **Automatic Detection**: Distinguishes between known and custom models
- **Future-Proof**: No need to update the code when new models are released

### **Provider-Specific Models**
- **OpenAI**: gpt-4o, gpt-4o-mini, gpt-4-turbo, gpt-3.5-turbo, etc.
- **Anthropic**: claude-3-5-sonnet, claude-3-haiku, claude-3-opus, etc.
- **Google**: gemini-1.5-pro, gemini-1.5-flash, gemini-pro, etc.

### **Model Commands**
```bash
# List all available models (known + custom)
./mad config model list

# Set any model (known or custom)
./mad config model set gpt-4o                    # Known model
./mad config model set my-custom-model           # Custom model

# Query live API for current models
./mad config model refresh                       # Get live model list
```

### **How It Works**
1. **Known Models**: Curated list of popular models (updated periodically)
2. **Custom Models**: Any model name you specify - the system attempts to use it
3. **Live API Queries**: `refresh` command queries provider APIs for current models
4. **Smart Validation**: Warns about custom models but doesn't block them
5. **Fallback Support**: Uses known models when API is unavailable
6. **Provider Switching**: Model selection is remembered per provider

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

### **Model Updates**
When contributing model updates:
- Update the `getKnownModels()` function in `cmd/config.go`
- Test with both known and custom model names
- Ensure provider switching works correctly

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Built with [Cobra CLI](https://cobra.dev/) for the command-line interface
- Uses [Google GenAI](https://ai.google.dev/) for Gemini integration
- Mermaid diagram syntax validation and rendering
- Flexible model management inspired by modern AI tooling patterns
