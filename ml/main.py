from fastapi import FastAPI, UploadFile, File, HTTPException
import uvicorn
from service.language_detection import detect_language_stub
import io

app = FastAPI(title="Voice Agent ML Sidecar")

@app.get("/health")
def health_check():
    return {"status": "ok"}

@app.post("/api/v1/detect-language")
async def detect_language(audio: UploadFile = File(...)):
    if not audio.filename:
        raise HTTPException(status_code=400, detail="No file uploaded")
    
    # Read the PCM audio data
    contents = await audio.read()
    
    # Call stub logic for MVP
    result = detect_language_stub(contents)
    
    return result

if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=8000, reload=True)
