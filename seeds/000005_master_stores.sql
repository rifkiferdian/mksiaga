-- Master store awal. ID dipertahankan agar kompatibel dengan data sumber.
START TRANSACTION;

INSERT INTO stores (id, code, name, address, timezone, status, deleted_at) VALUES
(1,   'MK1',  'MK1 Babarsari',          'Babarsari',   'Asia/Jakarta', 'active', NULL),
(2,   'MK2',  'MK2 Simanjuntak',        'Simanjuntak', 'Asia/Jakarta', 'active', NULL),
(3,   'MK3',  'MK3 Supeno',             'Supeno',      'Asia/Jakarta', 'active', NULL),
(4,   'MK4',  'MK4 Palagan',            'Palagan',     'Asia/Jakarta', 'active', NULL),
(5,   'MK5',  'MK5 Godean',             'Godean',      'Asia/Jakarta', 'active', NULL),
(6,   'MK6',  'MK6 Imogiri',            'Imogiri',     'Asia/Jakarta', 'active', NULL),
(7,   'MK7',  'MK7 Keloran',            'Keloran',     'Asia/Jakarta', 'active', NULL),
(101, 'MKM1', 'MK Mini 1 Plemesewu',    'Plemesewu',   'Asia/Jakarta', 'active', NULL),
(102, 'MKM2', 'MK Mini 2 Diro',         'Diro',        'Asia/Jakarta', 'active', NULL),
(103, 'MKM3', 'MK Mini 3 Minomartani',  'Minomartani', 'Asia/Jakarta', 'active', NULL)
ON DUPLICATE KEY UPDATE
    code = VALUES(code),
    name = VALUES(name),
    address = VALUES(address),
    timezone = VALUES(timezone),
    status = VALUES(status),
    deleted_at = NULL;

COMMIT;
