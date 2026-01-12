# 🤖 Learnloop

Learnloop is a framework for AI-powered coding tools (like Windsurf) that prevents your projects from turning into unmanageable "black boxes." It's a simple system of rules, prompts, and workflows designed to:


1. **Add Structure:** Guide your AI towards consistent, maintainable practices.
2. **Preserve Context:** Automatically save plans, progress, insights, and learnings.
3. **Empower You:** Keep you in control and ensure you *understand* the code being built.

**Because "vibecoding" is fun... until it’s not.**

Ever felt that thrill of AI generating code at lightning speed, only to hit a wall later, staring at a complex system you barely understand? Lost track of decisions? Can't remember why a certain approach was taken? That's the chaos Learnloop is built to prevent. We aim for *structured collaboration* with your AI, not just blind code generation.

**Is Learnloop Worth Your Time?**

If you're using AI to build something non-trivial – whether you're a seasoned dev, a newcomer leveraging AI, a designer, or a product person bringing ideas to life – and you need to:

* **Understand** the resulting codebase.
* **Maintain & Scale** it over time.
* **Retain control** throughout the development process.
* **Capture valuable learnings** along the way.

...then yes, Learnloop is designed for you. It helps bridge the gap between AI's speed and the need for sustainable, understandable software.

## Getting Started (2min)

```bash
git clone https://github.com/untoldecay/LearnLoop-AI-Prompt-Framework.git learnloop
```


1. **Add Framework Files:** Clone or download & copy the `_rules` folder to your project's root directory.
2. **Configure AI:** Copy the *entire content* of `_rules/.windsurfrules` into your AI tool's "custom instructions" or equivalent settings. This file contains the core operating principles.
3. **Declare Mission:** Edit `_rules/mission.md` with your project's high-level goal. This gives the AI crucial context.

That's it! Your AI assistant is now enhanced with Learnloop's structure.

## Demo

(Coming ASAP)

## How To Use Learnloop (Quick Overview)

Learnloop enhances your AI interaction through **passive background rules** and **active @ commands** you issue for specific tasks.

* **Passive:** The AI automatically tries to follow best practices defined in .windsurfrules which are loaded in its settings([see detailed explanations here](#the-ruleset-(.windsurfrules))).
* **Active Example:** Take direct control for structured tasks like planning a feature:

  > > "Please follow @breakdown.md to create a plan for the login form"

Explore specific commands in **[Prompts](#the-prompts-(_rules/_prompts/))** section.

# How Does It Work Under the Hood?

Learnloop relies on a few key components working together:

## Passive Guidance & Active Commands Explained

Learnloop works in two complementary ways: automatically in the background (passive) and when you explicitly command it (active).

### 1. Passive Mode: Your AI Follows the Rules Automatically

Once you've loaded `.windsurfrules` into your AI's custom instructions, it acts as a set of **core directives**. Your AI will automatically try to adhere to these rules during your conversation.

* **How it works:** The ruleset (`.windsurfrules`) tells the AI how to behave in certain situations (e.g., planning before coding, ensuring tests, using structured commits). It might also passively trigger specific prompts for common tasks based on your requests.
* **What you'll see:** The AI might state which rule it's applying, ask clarifying questions before generating complex code, suggest creating tests, or automatically format its output for readability, all *without* you explicitly telling it to use a Learnloop command. It aims to enforce good habits by default.
* **Example Trigger:** If you ask generally "Can you build the user profile section?", the rules might passively guide the AI to first ask about requirements or suggest using the `breakdown.md` process, referencing the rule it's following.

### 2. Active Mode: You Take Direct Command with `@` Prompts

For specific, structured tasks, you can **explicitly tell the AI** to use a Learnloop prompt template using the `@` symbol followed by the prompt filename.

* **How it works:** You invoke a predefined workflow template stored in the `_rules/_prompts/` folder. This ensures consistent execution of complex tasks like planning, refactoring, or documentation.
* **What you'll see:** The AI acknowledges the specific prompt command and follows the detailed steps outlined within that prompt's template file.
* **Example Usage:**

  **Instead of (vague):**

  > "Can you write code for a login form based on our conversation?"

  **Try (structured Learnloop command):**

  > "Please follow `@breakdown.md` to create a plan for a login form based on our findings in `@_rules/_analysis/[topic]/`"

  **The AI Will (following** `@breakdown.md`):

  
  1. Use context from previous analysis (if specified).
  2. Generate a structured, trackable implementation plan (`_rules/_features/.../plan.md`).
  3. Set up logging for implementation insights (`devlog` folder).
  4. Guide you step-by-step, potentially logging learnings (`_rules/_learnings/ai/`).
  5. Prompt for tests (likely in `_tests/`).
  6. Seek validation throughout the process.

This active mode gives you precise control over *when* and *how* specific structured workflows are executed.

## The Ruleset (.windsurfrules)

This is the **brain** or **operating system** loaded into your AI's persistent memory (custom instructions). It contains the core principles and conditional logic. It acts like an index, telling the AI *when* to apply general good practices or *when* to potentially trigger a specific prompt passively.

*Extract from* `_rules/.windsurfrules`:

```
    [...]
    * IF USER ASK TO BUILD A FEATURE or a new feature:
    * Search for the feature in '_rules/_requirements/' (for context, PR are not mandatory)
    * Verify if the feature is started already inside '_rules/_features/'. If true, continue based on the ETA and plan found.
    * If false, refer to '_rules/_prompts/breakdown.md' to create a plan for this feature. # <-- Passive Trigger Example
    [...]
```

**Key Rules Enforced (Passively):**

* **Readability:** Debug logs & comments.
* **Transparency:** State rules applied in output.
* **Documentation:** Encourage docstrings, params, returns.
* **Planning:** Plan before coding, break down tasks.
* **Error Handling:** Use try-except, meaningful messages.
* **Testing:** Promote self-contained tests in `_tests/`.
* **Security:** Avoid hardcoded secrets, use env vars.
* **Maintainability:** Promote focused functions, DRY, clean architecture.
* **Modularity:** Suggest Facade pattern for large files.
* **Version Control:** Guide structured commit messages.

## The Prompts (_rules/_prompts/)

These are detailed **workflow templates** or **Standard Operating Procedures (SOPs)** for specific tasks. They ensure consistency and automation.

* **Triggered Passively:** By conditions met in the `.windsurfrules`.
* **Triggered Actively:** By you using the `@prompt_name.md` command in chat.

### Featured Prompts (Examples of Active Usage):

#### Create a Feature Plan (`@breakdown.md`)

* **Goal:** Start building a new feature in a structured way.
* **Try:** `> "follow @breakdown.md for creating a [feature-name]"`
* **Outcome:** Analyzes context, creates a detailed plan in `_rules/_features/[feature-name]/plan.md`, sets up `devlog`, saves insights, guides implementation.

#### Breaking Down Large Files (`@facade_modularization.md`)

* **Goal:** Refactor large, unwieldy files into manageable modules.
* **Try:** `> "use @facade_modularization.md on @[file-or-folder-name]"`
* **Outcome:** Analyzes file(s), proposes a Facade pattern modularization plan, guides refactoring step-by-step.

#### Creating Learning Cards (`@add-user-learnings.md`)

* **Goal:** Capture and solidify important concepts or insights YOU gained.
* **Try:** `> "use @add-user-learnings.md to create a learning card about [topic we discussed]"`
* **Outcome:** Analyzes conversation, extracts key points, creates a structured learning card in `_rules/_learnings/user/`.

#### Creating Documentation (`@documentation.md`)

* **Goal:** Document a completed feature or component consistently.
* **Try:** `> "create @documentation.md for this component @[file-or-folder-name]"`
* **Outcome:** Analyzes code, generates standardized documentation (descriptions, params, examples) in `_rules/_documentation/`.

#### Teaching Mode

* **Goal:** Learn a new technology with guided assistance, focusing on understanding.
* **Try:** `> "enter teaching mode for learning @[technology-name]"`
* **Outcome:** Creates a learning plan, guides *you* through implementation, provides feedback, explains concepts.

*(More prompts like analysis capture and context extraction can be added.)*

## The Folder Structure (`_rules/` & `_tests/`)

This is the **persistent project memory** or **shared brain**. Learnloop uses a defined structure to save the outputs generated by the rules and prompts, making context accessible across sessions for both you and the AI.

```
_rules/                  # Learnloop's structured context home
 ├── .windsurfrules      # AI Core Directives (Loaded in AI memory)
 ├── .mission            # Project Goal Context
 ├── _prompts/           # Workflow Templates (SOPs)
 ├── _features/          # Feature Plans & Devlogs (Generated Output)
 ├── _documentation/     # Feature/Component Docs (Generated Output)
 └── _learnings/         # User & AI Insights (Generated Output)
     ├── user/
     └── ai/
_tests/                  # Self-contained tests (Outside _rules, but encouraged by rules)
```

Each interaction guided by Learnloop helps build this structured knowledge base, improving future results and understanding.

## Benefits: Build Faster, Smarter, and Stay in Control

Learnloop helps you move beyond chaotic AI code generation towards structured, sustainable development:

| Benefit | Description |
|:---|:---|
| **Regain Control & Understanding** | Break down complexity, follow clear plans, understand *how* and *why* code is built. Avoid the black box! |
| **Persistent Project Brain** | Context, plans, and insights are saved systematically in `_rules/`, not lost in chat history. |
| **Working Continuity** | Track feature progress (`_features/`) across sessions, easily pick up where you left off. |
| **Capture User Learnings** | Actively solidify your own understanding with structured learning cards (`_learnings/user/`). |
| **Structured Documentation** | Automatically generate consistent, accessible documentation (`_documentation/`). |
| **Enforced Best Practices** | Ruleset passively encourages testing, readability, modularity etc., improving code quality. |
| **Structured Context for AI** | Organized context in `_rules/` helps the AI understand the project better over time. |
| **Tool Interoperability** | The structured text-based context in `_rules/` can potentially be referenced by other AI tools. |
| **Focused AI Interaction** | Reference specific context files (`@_rules/...`) to save tokens and guide the AI more precisely. |
| **Knowledge Sharing Ready** | Structured outputs (plans, docs, learnings) are easier to share with team members. |

Learnloop bridges the gap between rapid AI code generation and sustainable development practices. Build working software *that you fully comprehend and can confidently maintain*.

## About me

I'm product designer, illustrator, and front end developer (jquery/php coughing dust clouds), and AI enthousiast apparently. Bridging ideas and code is my jam, and I believe AI should empower that, not obscure it.

Thanks to [Yifan](https://github.com/yifan) for having opened this window. So much useful insights on AI development. *(my first breakdown prompt version was his, and I still have his versionning settings inside .windsurfrules I believe)*

Thanks to [Guiz](https://github.com/guillem-cotcha) for the idea of AI context enforcement by logging success and fails over feature implementation.


---

*Made with ❤️ for people who want control & understanding in their AI coding journey.* By [@untoldecay](https://x.com/untoldecay)


---

