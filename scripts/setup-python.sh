#!/bin/bash

# Python setup for ML predictor components (future enhancement)

echo "Setting up Python environment for GPU Scheduler ML components..."

# Activate venv if it exists
if [ -d "venv" ]; then
    echo "Activating existing virtual environment..."
    source venv/bin/activate
else
    echo "Creating new virtual environment..."
    python3.12 -m venv venv
    source venv/bin/activate
fi

echo "Virtual environment activated: $(which python)"
echo "Python version: $(python --version)"

# Install Python dependencies for future ML features
echo ""
echo "Installing Python packages for ML prediction..."

pip install --upgrade pip

# Core ML libraries
pip install numpy pandas scikit-learn

# XGBoost for job completion prediction
pip install xgboost lightgbm

# Plotting and analysis
pip install matplotlib seaborn

# API client
pip install requests

echo ""
echo "✓ Python environment setup complete!"
echo ""
echo "To use this environment:"
echo "  source venv/bin/activate"
echo ""
echo "Installed packages:"
pip list
echo ""
echo "This Python environment will be used for:"
echo "  - Job completion time prediction (ML model)"
echo "  - Feature extraction from job metadata"
echo "  - Model training and evaluation"
echo "  - Data analysis and visualization"
echo ""
