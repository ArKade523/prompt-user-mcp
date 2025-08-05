# User input MCP server

This is an MCP server that is intended to increase interaction between a user and an agentic LLM tool by giving the LLM the option to ask for feedback without stopping its train of thought.

It has two primary use cases:

1.  Asking the user to test a feature and provide feedback.
    For example:
    - Asking the user to observe and test a UI. 
    - Asking the user to test a tool with a sensitive API key.
    - Asking the user to build and flash an embedded program and report behavior.

2. Waiting for approval from the user before continuing implementation.
    For example:
    - The user could require manual approval of each feature before continuing implementation. This is important for keeping development on track, instead of wasting time and resources on a slightly wrong application.
    - The user could ask for an additional feature set before continuing with a prior plan.
    - The user could ask the agent to use a different API before getting further entrenched in a specific ecosystem.
    - The user could ask the agent to switch to a different set of libraries, framework, or language.
    
