CREATE TABLE IF NOT EXISTS prestasi (
    id_prestasi     SERIAL          PRIMARY KEY,
    id_student      INT             NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    nama_prestasi   VARCHAR(255)    NOT NULL,
    juara           INT             NOT NULL
);

insert into prestasi (id_student, nama_prestasi, juara) values (2, 'osn', 3);
insert into prestasi (id_student, nama_prestasi, juara) values (2, 'osn-k', 1);
