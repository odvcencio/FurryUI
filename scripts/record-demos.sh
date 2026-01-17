#!/bin/bash
# Record demo screencasts for all FluffyUI examples
#
# Usage: ./scripts/record-demos.sh
#
# Requirements:
#   - Go 1.25+
#   - Optional: agg (for GIF conversion)
#   - Optional: ffmpeg (for MP4 conversion)
#
# This script records terminal sessions as .cast files (Asciicast v2 format).
# These can be viewed with asciinema or converted to GIF/video.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
DEMOS_DIR="$ROOT_DIR/docs/demos"

# Duration to record each demo (seconds)
DURATION=${DURATION:-5}

# Ensure demos directory exists
mkdir -p "$DEMOS_DIR"

# Color output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}FluffyUI Demo Recorder${NC}"
echo "Recording demos to: $DEMOS_DIR"
echo "Duration per demo: ${DURATION}s"
echo ""

# Function to record a single example
record_example() {
    local name=$1
    local path=$2
    local cast_file="$DEMOS_DIR/${name}.cast"

    echo -e "${YELLOW}Recording: $name${NC}"

    # Run the example with recording enabled, auto-terminate after DURATION seconds
    # Using timeout to limit execution time
    timeout --signal=SIGINT "${DURATION}s" \
        env FLUFFYUI_RECORD="$cast_file" \
        go run "$path" 2>/dev/null || true

    if [ -f "$cast_file" ]; then
        echo "  -> Created: $cast_file"

        # Convert to GIF if agg is available
        if command -v agg &> /dev/null; then
            local gif_file="$DEMOS_DIR/${name}.gif"
            echo "  -> Converting to GIF..."
            agg --theme monokai --font-size 14 "$cast_file" "$gif_file" 2>/dev/null && \
                echo "  -> Created: $gif_file" || \
                echo "  -> GIF conversion failed"
        fi
    else
        echo "  -> Recording failed (no output file)"
    fi
    echo ""
}

# Record each example
# Note: Some examples require interactive input, so they may not record well
# The simpler examples work best for demos

echo "=== Recording Simple Examples ==="
record_example "quickstart" "$ROOT_DIR/examples/quickstart"
record_example "counter" "$ROOT_DIR/examples/counter"

echo "=== Recording Feature Examples ==="
record_example "todo-app" "$ROOT_DIR/examples/todo-app"
record_example "command-palette" "$ROOT_DIR/examples/command-palette"
record_example "settings-form" "$ROOT_DIR/examples/settings-form"
record_example "dashboard" "$ROOT_DIR/examples/dashboard"

echo "=== Recording Showcase ==="
record_example "candy-wars" "$ROOT_DIR/examples/candy-wars"

echo "=== Recording Widget Galleries ==="
record_example "widgets-gallery" "$ROOT_DIR/examples/widgets/gallery"
record_example "widgets-layout" "$ROOT_DIR/examples/widgets/layout"
record_example "widgets-input" "$ROOT_DIR/examples/widgets/input"
record_example "widgets-data" "$ROOT_DIR/examples/widgets/data"
record_example "widgets-navigation" "$ROOT_DIR/examples/widgets/navigation"
record_example "widgets-feedback" "$ROOT_DIR/examples/widgets/feedback"

echo -e "${GREEN}Recording complete!${NC}"
echo ""
echo "Files created in: $DEMOS_DIR"
echo ""
echo "To view recordings:"
echo "  asciinema play $DEMOS_DIR/quickstart.cast"
echo ""
echo "To convert to GIF (requires agg):"
echo "  agg --theme monokai $DEMOS_DIR/quickstart.cast $DEMOS_DIR/quickstart.gif"
echo ""
echo "To convert to MP4 (requires agg + ffmpeg):"
echo "  agg $DEMOS_DIR/quickstart.cast /tmp/quickstart.webm"
echo "  ffmpeg -i /tmp/quickstart.webm $DEMOS_DIR/quickstart.mp4"
