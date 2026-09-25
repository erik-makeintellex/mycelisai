-- Migration 041 down: remove operator templates (RC-004G clones).

DROP TABLE IF EXISTS operator_templates;
