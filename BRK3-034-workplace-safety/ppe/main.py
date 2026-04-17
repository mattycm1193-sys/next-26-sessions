"""
 Copyright 2026 Google LLC

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
 """

import base64
import logging
import os
import json
import vertexai
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from vertexai.generative_models import GenerativeModel, Part, GenerationConfig

# Infrastructure Config
PROJECT_ID = os.environ.get("PROJECT_ID")
LOCATION = os.environ.get("LOCATION")

vertexai.init(project=PROJECT_ID, location=LOCATION)

logging.basicConfig(level=logging.INFO, format="%(asctime)s | %(levelname)s | %(message)s")
logger = logging.getLogger(__name__)

# --- NEW: System Instructions for High Precision ---
SYSTEM_INSTRUCTION = """
You are a high-precision Industrial Safety Inspector. Your task is to perform an exhaustive audit of images for PPE compliance.
1. Accuracy is critical. Distinguish clearly between:
   - 'hard_hat': Rigid, plastic safety helmets (yellow, white, orange, etc.) used in construction.
   - 'regular_hat': Soft fabric hats, baseball caps, beanies, or hoods. These are NOT PPE.
2. Be exhaustive. Detect EVERY person and EVERY piece of headwear, even if they are small or in the background.
3. Return ONLY a valid JSON object.
"""

app = FastAPI()
app.add_middleware(CORSMiddleware, allow_origins=["*"], allow_methods=["*"], allow_headers=["*"])

class ImageData(BaseModel):
    image_base64: str

# Initialize model with System Instructions
model = GenerativeModel(
    "gemini-2.5-flash",
    system_instruction=[SYSTEM_INSTRUCTION]
)

@app.post("/analyze")
async def analyze_frame(data: ImageData):
    try:
        image_bytes = base64.b64decode(data.image_base64)
        image_part = Part.from_data(data=image_bytes, mime_type="image/jpeg")

        # Refined Prompt for Specificity
        prompt = """
        Exhaustively detect all instances of the following:
        - 'person'
        - 'hard_hat' (specifically industrial safety helmets)
        - 'regular_hat' (non-safety hats like baseball caps)
        
        Return JSON: {"objects": [{"label": "...", "box_2d": [ymin, xmin, ymax, xmax]}]}
        Use coordinates 0-1000.
        """

        response = model.generate_content(
            [image_part, prompt],
            generation_config=GenerationConfig(
                response_mime_type="application/json",
                temperature=0.1, # Lower temperature increases precision/consistency
            ),
        )

        raw_results = json.loads(response.text)
        objects = raw_results.get("objects", [])
        
        formatted_detections = []
        person_found = False
        ppe_found = False

        for obj in objects:
            label = obj['label'].lower()
            ymin, xmin, ymax, xmax = obj['box_2d']
            
            box = [
                {"x": xmin/1000, "y": ymin/1000},
                {"x": xmax/1000, "y": ymin/1000},
                {"x": xmax/1000, "y": ymax/1000},
                {"x": xmin/1000, "y": ymax/1000}
            ]

            # STRICT PPE CHECK
            is_ppe = label in ["hard_hat", "helmet", "safety helmet"]
            if label == "person": person_found = True
            if is_ppe: ppe_found = True

            formatted_detections.append({
                "label": "Hard Hat" if is_ppe else label.title(),
                "box": box,
                "is_ppe": is_ppe
            })

        ppe_ok = person_found and ppe_found
        message = "PPE VERIFIED" if ppe_ok else "SAFETY VIOLATION"
        if not person_found: message = "NO PERSON DETECTED"
        if person_found and not ppe_found: message = "NO HARD HAT DETECTED"

        return {"ppe_ok": ppe_ok, "message": message, "detections": formatted_detections}

    except Exception as e:
        logger.error(f"Analysis Failed: {str(e)}")
        return {"ppe_ok": False, "message": "Detection Error", "detections": []}
