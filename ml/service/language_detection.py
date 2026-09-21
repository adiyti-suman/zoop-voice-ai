import random
from typing import Dict, Any

def detect_language_stub(audio_data: bytes) -> Dict[str, Any]:
    """
    Mock stub for Language Detection.
    In Phase 06 MVP, we'll return a high probability for a random Indian language
    or English, to simulate the LID model output without loading heavy torch models yet.
    """
    
    # Check if there is actual audio data
    is_speech = len(audio_data) > 1000
    
    if not is_speech:
        return {
            "probabilities": [],
            "is_speech": False
        }
        
    # Simulated detected languages (English, Hindi, Kannada, Tamil)
    languages = ["en", "hi", "kn", "ta"]
    top_lang = random.choice(languages)
    
    # Generate fake probabilities
    probs = [
        {"language": top_lang, "score": 0.85 + random.random() * 0.1},
        {"language": random.choice([l for l in languages if l != top_lang]), "score": 0.05},
        {"language": random.choice([l for l in languages if l != top_lang]), "score": 0.01}
    ]
    
    # Sort by score descending
    probs.sort(key=lambda x: x["score"], reverse=True)
    
    return {
        "probabilities": probs,
        "is_speech": True
    }
