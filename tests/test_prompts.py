import sys
import os
# Add project root to sys.path
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from agents.prompts import build_system_prompt

def test_build_system_prompt():
    name = "TestAgent"
    team = "TestTeam"
    capabilities = ["tool_a", "tool_b"]
    custom_instructions = "Do the thing."
    
    assignment = "Data Ingestor"
    
    prompt = build_system_prompt(name, team, capabilities, custom_instructions, assignment)
    
    assert "You are TestAgent" in prompt
    assert "team: TestTeam" in prompt
    assert "**Assignment**: Data Ingestor" in prompt
    assert "tool_a, tool_b" in prompt
    assert "TOOL: tool_name" in prompt # Check for syntax instruction
    assert "Do the thing." in prompt
    assert "# GOVERNANCE & PROTOCOLS" in prompt

def test_build_system_prompt_no_tools():
    prompt = build_system_prompt("Agent", "Team", [], "Instructions")
    assert "None" in prompt # Should say capabilities: None
