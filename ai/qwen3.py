import torch
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from transformers import AutoTokenizer, AutoModelForCausalLM

# 定义输入数据模型
class RequestBody(BaseModel):
    prompt: str

app = FastAPI()

# 全局变量，存储模型和分词器
tokenizer = None
model = None

# 加载模型
@app.on_event("startup")
def load_model():
    global tokenizer, model
    try:
        model_path = "E:\hugging_face_models\Qwen\Qwen3-0___6B"
        tokenizer = AutoTokenizer.from_pretrained(model_path, trust_remote_code=True)
        model = AutoModelForCausalLM.from_pretrained(
            model_path,
            trust_remote_code=True,
            device_map="auto" # GPU
        ).eval()
        print("模型加载成功！")
    except Exception as e:
        print(f"模型加载失败: {e}")
        raise RuntimeError("模型加载失败，请检查路径和文件。")


@app.post("/generate")
async def generate_text(body: RequestBody):
    # 检查
    if not model or not tokenizer:
        raise HTTPException(status_code=503, detail="模型未加载或加载失败。")

    try:
        messages = [
            {"role": "user", "content": body.prompt},
        ]

        # 历史
        input_ids = tokenizer.apply_chat_template(
            messages,
            tokenize=True,
            add_generation_prompt=True,
            return_tensors="pt"
        )

        input_ids = input_ids.to(model.device)

        # 记录输入ID的长度
        input_length = input_ids.shape[1]

        # 模型生成
        outputs = model.generate(
            input_ids,
            max_new_tokens=1024,
            do_sample=True,  # 采样策略
            temperature=0.8,  # 采样温度
            top_p=0.9,
            repetition_penalty=1.1
        )

        # 4. 解码生成的文本
        response_ids = outputs[0][input_length:]
        clean_response = tokenizer.decode(response_ids, skip_special_tokens=True)

        print(f"成功生成响应: {clean_response}")
        return {"response": clean_response}
    except Exception as e:
        print(f"推理失败: {e}")
        raise HTTPException(status_code=500, detail=f"模型推理失败: {str(e)}")
