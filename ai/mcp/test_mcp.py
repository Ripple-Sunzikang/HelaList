from typing import Any, List, Dict, Optional
import asyncio
import json
import os
import pandas as pd

from fastmcp import FastMCP

mcp = FastMCP("mcp_test")

@mcp.tool()
async def hello_world(name: Optional[str] = None) -> str:
    """输出helloworld消息，支持自定义名称
    
    Args:
        name: 可选参数，自定义名称（如传入则输出'hello, {name}!'）
    """
    # 如果传入了名称参数，则个性化问候
    if name:
        return f"hello, {name}! 这是MCP工具输出的消息"
    # 未传入参数则输出默认helloworld
    return "helloworld! 这是MCP工具输出的消息"

if __name__ == "__main__":
    # 初始化并运行服务器

    mcp.run(transport='stdio')
