bd devlog record — Feedback Report                                                       
                                                                                           
  Observation                                                                              
                                                                                           
  bd devlog record --subject "..." --problem "..." --file "filename.md" is misleading about
   what it does vs what agents expect it to do.                                            
   
  What it does                                                                             
                                                                                           
  - Registers a session entry in the devlog index (_index.md)                              
  - Points to an existing markdown file via --file                                         
  - Syncs the DB
  - Does NOT validate or inspect the file content                                          
                                              
  What agents (and humans) expect                                                          
                                                                                           
  - That record captures the current session's work into a devlog                          
  - That the --subject and --problem flags describe the content that will be in the file
                                                                                           
  The problem in practice                                                                  
                                          
  During this session, I ran bd devlog record twice pointing to the same file              
  (2026-05-29_fix-xpost-video-pipeline.md):                                                
                                                                                           
  1. First call (May 29): File existed with matching content — correct.                    
  2. Second call (May 30): New subject/problem, but --file pointed to the same May 29 file.
   The record succeeded, creating a new session entry in the index — but the file content  
  was stale (yesterday's work, not today's). No warning was emitted.                       
   
  I didn't realize the devlog for today's work was missing until the user asked. The record
   command gave a green checkmark both times.                                              
                                                                                           
  Suggestions                                                                              
   
  1. Warn on content mismatch — If the --subject text doesn't appear anywhere in the       
  referenced file, warn: "Subject doesn't match file content — did you mean to create a new
   file?"
  2. Warn on reuse — If the same --file is already registered to another session, warn:    
  "This file is already linked to session sess-XXXX. Create a new file?"
  3. Auto-stub option — The v0.52.0 changelog mentions "Atomic Record — bd devlog record   
  auto-creates markdown stubs with AI directives." If this is implemented, make it the     
  default behavior when the file doesn't exist, and document it prominently. During this   
  session it wasn't clear whether --file should point to an existing file or would create
  one.                                                                                     
  4. Rename or alias — Consider bd devlog register for the current behavior (linking an    
  existing file) and bd devlog record for the full workflow (create file + register). The  
  word "record" implies capturing, not just linking.                                       

  Environment                                                                              
   
  - bd v0.52.0                                                                             
  - Agent: Claude Opus 4.6 via Claude Code                                                 
  - The CLAUDE.md protocol says to run bd devlog record after writing the file, but the
  command's UX doesn't enforce that sequence
