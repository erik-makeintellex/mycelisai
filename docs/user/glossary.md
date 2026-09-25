# Glossary

> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

## Organization

An instantiated AI Organization is a dedicated workspace where you work with Soma and teams. Each organization has its own settings, teams, approval policies, and retained outputs.

## Template

A reusable blueprint for organizing teams, assigning roles, setting approval policies, and defining which AI models and tools are available. Selecting a template when you create an organization sets up defaults, but editing a template never changes organizations that were already created from it.

## Team

A named set of agents with a lead who coordinates work. Standing teams load automatically when the system starts and remain available for ongoing work. Mission teams are created temporarily for a specific task and can be archived when the mission is complete.

## Group

A governed work scope that defines how decisions are made and what authority agents have. A group has a work mode (such as "propose only" or "execute with approval"), an approval policy that describes who must review actions, and optionally an expiration date. Groups can include teams and people.

## Deployment

Anything currently defined to run in your organization: a standing team, a mission team, a group, or a planned department. The [Deployments](deployments.md) page (`/deployments`) lists every deployment, shows their state (running, idle, planned, paused, archived, expired, or error), and explains what each one does, who is on it, and what it produced.

## Knowledge

Documents, notes, data, and context that Soma can cite when answering questions or making decisions. Knowledge is stored separately so you can organize private records, customer context, company guidance, reflection notes, and other information by sensitivity and purpose. Formerly called "Deployment Context."
