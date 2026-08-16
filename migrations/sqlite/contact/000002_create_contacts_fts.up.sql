CREATE VIRTUAL TABLE IF NOT EXISTS contacts_fts USING fts5(
    name,
    emails,
    phones,
    addresses,
    notes,
    content='contacts',
    content_rowid='id'
);

CREATE TRIGGER IF NOT EXISTS contacts_ai AFTER INSERT ON contacts BEGIN
    INSERT INTO contacts_fts(rowid, name, emails, phones, addresses, notes)
    VALUES (new.id, new.name, new.emails, new.phones, new.addresses, new.notes);
END;

CREATE TRIGGER IF NOT EXISTS contacts_ad AFTER DELETE ON contacts BEGIN
    INSERT INTO contacts_fts(contacts_fts, rowid, name, emails, phones, addresses, notes)
    VALUES ('delete', old.id, old.name, old.emails, old.phones, old.addresses, old.notes);
END;

CREATE TRIGGER IF NOT EXISTS contacts_au AFTER UPDATE ON contacts BEGIN
    INSERT INTO contacts_fts(contacts_fts, rowid, name, emails, phones, addresses, notes)
    VALUES ('delete', old.id, old.name, old.emails, old.phones, old.addresses, old.notes);
    INSERT INTO contacts_fts(rowid, name, emails, phones, addresses, notes)
    VALUES (new.id, new.name, new.emails, new.phones, new.addresses, new.notes);
END;
