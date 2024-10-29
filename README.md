# Computer Use Project

This project consists of multiple components working together to provide a desktop environment with various services and applications. The main components include:

1. Desktop Environment Services
2. Computer Use Demo Application
3. Text Adventure Game Services

## Prerequisites

- Ubuntu/Linux environment
- Python 3.11+
- Go 1.x+
- Node.js (for noVNC)

## Project Structure

```
.
├── computer_use_demo/         # Demo application with Streamlit interface
├── text_adventure_game/       # Text adventure game services
├── static_content/            # Static web content
├── *.sh                       # Various startup scripts
└── http_server.py            # HTTP server implementation
```

## Installation

1. Install system dependencies:
```bash
sudo apt update
sudo apt install -y python3-pip golang-go nodejs npm xvfb x11vnc novnc
```

2. Install Python dependencies:
```bash
# For computer_use_demo
cd computer_use_demo
pip install -r requirements.txt

# For text adventure services
cd ../text_adventure_game/textadventureservices
pip install -r requirements.txt
```

3. Install Go dependencies:
```bash
cd text_adventure_game/go-services
go mod download
```

## Starting the Services

The project uses several startup scripts to initialize all necessary services. The main startup script `start_all.sh` orchestrates the launch of all components.

To start all services:

```bash
./start_all.sh
```

This script will start the following components in sequence:

1. Xvfb (X Virtual Frame Buffer) - `xvfb_startup.sh`
2. Mutter (Window Manager) - `mutter_startup.sh`
3. Tint2 (Panel/Taskbar) - `tint2_startup.sh`
4. x11vnc (VNC Server) - `x11vnc_startup.sh`
5. noVNC (VNC Client) - `novnc_startup.sh`

### Individual Service Details

#### Desktop Environment
- **Xvfb**: Virtual X server, runs on display :1
- **Mutter**: Window manager for the desktop environment
- **Tint2**: Lightweight panel/taskbar
- **x11vnc**: VNC server that shares the Xvfb display
- **noVNC**: HTML5 VNC client accessible via web browser

#### Computer Use Demo
The demo application is a Streamlit-based interface that provides various tools and functionalities:
```bash
cd computer_use_demo
streamlit run streamlit.py
```

#### Text Adventure Game Services
The text adventure game consists of multiple microservices:
```bash
cd text_adventure_game/go-services
# Start each service in its own terminal:
cd services/auth && go run .
cd services/gamestate && go run .
cd services/worldgen && go run .
```

## Accessing the Services

- **Desktop Environment**: Access through noVNC at `http://localhost:6080/vnc.html`
- **Computer Use Demo**: Access the Streamlit interface at `http://localhost:8501`
- **Game Services**: Various endpoints available based on service specifications in `text_adventure_game/textadventureservices/Specs/`

## Configuration

- Environment variables and configuration settings can be found in:
  - `computer_use_demo/tools/config.py`
  - `text_adventure_game/textadventureservices/src/config/settings.py`

## Development

- Python code follows PEP 8 style guide
- Go code follows standard Go formatting conventions
- Use `go fmt` before committing Go code
- Run tests using `go test ./...` in Go service directories

## Troubleshooting

1. If the desktop environment doesn't start:
   - Check Xvfb logs
   - Ensure display :1 is not already in use
   - Verify all startup scripts have execute permissions

2. If services fail to start:
   - Check port availability
   - Verify all dependencies are installed
   - Check service logs in respective directories

3. If noVNC connection fails:
   - Verify VNC server is running
   - Check network connectivity
   - Ensure ports 6080 and 5901 are available

## License

This project is proprietary and confidential.

## Contributing

For internal use only. Please follow the established development workflow and code review process.