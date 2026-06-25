## Go Fyne GUI for stable-diffusion.cpp

A lightweight, cross-platform Graphical User Interface (GUI) for generating images with stable-diffusion.cpp, built in Go using the Fyne toolkit.
This project acts as a client frontend that connects directly to a running sd-server instance.

## 🚀 Features & Current State## Core Functionality

* txt2img & img2img: Full support for most stable-diffusion.cpp parameters.
* Inpainting: Full support using init_image, mask_image, and ref_images (control_image not tested).
* Mask Editor: Built-in simple graphic editor for drawing inpainting masks directly in the app.
* LoRA / LyCORIS: Controls for adding and adjusting model weights directly in the prompt.
* Upscaling: Built-in parameters for Hi-Res Fix and image upscaling.

## Gallery & Metadata

* File Manager: Simple built-in gallery for viewing generated images.
* EXIF Reader & Prompt History: Reads generation parameters directly from image metadata to quickly reuse them for new sessions.

------------------------------

## 🛠️ Prerequisites

To use this GUI, you must have an active instance of the sd-server (from the stable-diffusion.cpp project) running.

   1. Clone and build stable-diffusion.cpp.
   2. Start the server via terminal:

   ./sd-server -m /path/to/your/model.safetensors --listen-port 8080 --listen-ip 0.0.0.0

------------------------------

## 💻 Installation & Usage## 1. Clone the repository

git clone <https://github.com/olegk0/SD_UI.git>
cd SD_UI

## 2. Install dependencies

go mod tidy

Note: Fyne may require graphics development libraries depending on your OS (e.g., libgl1-mesa-dev on Linux or Xcode CLI tools on macOS).

## 3. Run the application

go run .

Inside the app, enter your sd-server URL/Port, configure your parameters, and start generating
------------------------------
