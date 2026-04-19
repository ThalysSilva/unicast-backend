-- Demo seed for TCC presentations.
--
-- Login:
--   email: demo@unicast.local
--   senha: Unicast@2026
--
-- Como executar:
--   psql "$POSTGRES_DATABASE_URL" -f scripts/demo-seed.sql
--
-- Ou, usando o container do docker-compose:
--   docker exec -i postgres-unicast psql -U "$POSTGRES_USER" -d unicast < scripts/demo-seed.sql
--
-- A seed é idempotente para os registros abaixo. Ela remove apenas o usuário
-- demo e as matrículas demo antes de recriar o cenário.

BEGIN;

DELETE FROM message_logs
WHERE student_id IN (
  '00000000-0000-4000-8000-000000004001',
  '00000000-0000-4000-8000-000000004002',
  '00000000-0000-4000-8000-000000004003',
  '00000000-0000-4000-8000-000000004004',
  '00000000-0000-4000-8000-000000004005',
  '00000000-0000-4000-8000-000000004006',
  '00000000-0000-4000-8000-000000004007',
  '00000000-0000-4000-8000-000000004008'
);

DELETE FROM students
WHERE student_id IN (
  '2026001',
  '2026002',
  '2026003',
  '2026004',
  '2026005',
  '2026006',
  '2026007',
  '2026008'
);

DELETE FROM users
WHERE id = '00000000-0000-4000-8000-000000000001'
   OR email = 'demo@unicast.local';

INSERT INTO users (id, email, name, password, salt, created_at, updated_at)
VALUES (
  '00000000-0000-4000-8000-000000000001',
  'demo@unicast.local',
  'Prof. Thalys Demo',
  '$2b$10$9umJAaMn/XV0bQwvXxXR/uDeeq7WI6Gs3hvZcEPS93.fjWYLD5GJm',
  decode('756e69636173742d64656d6f2d73616c74', 'hex'),
  now() - interval '30 days',
  now() - interval '30 days'
);

INSERT INTO campuses (id, name, description, user_owner_id, created_at, updated_at)
VALUES
  (
    '00000000-0000-4000-8000-000000001001',
    'Campus Centro',
    'Unidade principal usada para disciplinas presenciais e laboratórios.',
    '00000000-0000-4000-8000-000000000001',
    now() - interval '25 days',
    now() - interval '25 days'
  ),
  (
    '00000000-0000-4000-8000-000000001002',
    'Campus Norte',
    'Unidade de apoio para turmas noturnas e atividades de extensão.',
    '00000000-0000-4000-8000-000000000001',
    now() - interval '24 days',
    now() - interval '24 days'
  ),
  (
    '00000000-0000-4000-8000-000000001003',
    'Campus Leste',
    'Unidade dedicada a cursos híbridos, laboratórios maker e projetos de extensão.',
    '00000000-0000-4000-8000-000000000001',
    now() - interval '23 days',
    now() - interval '23 days'
  );

INSERT INTO programs (id, name, description, campus_id, active, created_at, updated_at)
VALUES
  (
    '00000000-0000-4000-8000-000000002001',
    'Ciência da Computação',
    'Curso de graduação com foco em desenvolvimento, dados e arquitetura de software.',
    '00000000-0000-4000-8000-000000001001',
    true,
    now() - interval '23 days',
    now() - interval '23 days'
  ),
  (
    '00000000-0000-4000-8000-000000002002',
    'Sistemas de Informação',
    'Curso voltado para sistemas corporativos, processos e gestão de tecnologia.',
    '00000000-0000-4000-8000-000000001001',
    true,
    now() - interval '22 days',
    now() - interval '22 days'
  ),
  (
    '00000000-0000-4000-8000-000000002003',
    'Análise e Desenvolvimento de Sistemas',
    'Curso tecnológico com turmas noturnas no Campus Norte.',
    '00000000-0000-4000-8000-000000001002',
    true,
    now() - interval '21 days',
    now() - interval '21 days'
  ),
  (
    '00000000-0000-4000-8000-000000002004',
    'Engenharia de Computação',
    'Curso com foco em sistemas embarcados, redes e integração hardware/software.',
    '00000000-0000-4000-8000-000000001003',
    true,
    now() - interval '20 days',
    now() - interval '20 days'
  ),
  (
    '00000000-0000-4000-8000-000000002005',
    'Gestão da Tecnologia da Informação',
    'Curso híbrido para gestão de serviços, governança e suporte de TI.',
    '00000000-0000-4000-8000-000000001003',
    true,
    now() - interval '19 days',
    now() - interval '19 days'
  ),
  (
    '00000000-0000-4000-8000-000000002006',
    'Ciência de Dados',
    'Curso experimental usado para demonstrar cursos ainda sem disciplinas ativas.',
    '00000000-0000-4000-8000-000000001002',
    true,
    now() - interval '18 days',
    now() - interval '18 days'
  );

INSERT INTO disciplines (id, name, description, year, semester, program_id, created_at, updated_at)
VALUES
  (
    '00000000-0000-4000-8000-000000003001',
    'Engenharia de Software',
    'Disciplina usada para demonstrar convites, auto-cadastro e envio segmentado.',
    2026,
    1,
    '00000000-0000-4000-8000-000000002001',
    now() - interval '20 days',
    now() - interval '20 days'
  ),
  (
    '00000000-0000-4000-8000-000000003002',
    'Banco de Dados II',
    'Turma com alunos de diferentes cursos para demonstrar filtros.',
    2026,
    1,
    '00000000-0000-4000-8000-000000002001',
    now() - interval '19 days',
    now() - interval '19 days'
  ),
  (
    '00000000-0000-4000-8000-000000003003',
    'Gestão de Projetos',
    'Disciplina compartilhada com Sistemas de Informação.',
    2026,
    1,
    '00000000-0000-4000-8000-000000002002',
    now() - interval '18 days',
    now() - interval '18 days'
  ),
  (
    '00000000-0000-4000-8000-000000003004',
    'Programação Web',
    'Turma noturna do Campus Norte.',
    2026,
    1,
    '00000000-0000-4000-8000-000000002003',
    now() - interval '17 days',
    now() - interval '17 days'
  ),
  (
    '00000000-0000-4000-8000-000000003005',
    'Tópicos Especiais em Integrações',
    'Disciplina sem alunos para demonstrar estados vazios.',
    2026,
    2,
    '00000000-0000-4000-8000-000000002003',
    now() - interval '16 days',
    now() - interval '16 days'
  ),
  (
    '00000000-0000-4000-8000-000000003006',
    'Arquitetura de Computadores',
    'Disciplina do Campus Leste com alunos compartilhados entre cursos.',
    2026,
    1,
    '00000000-0000-4000-8000-000000002004',
    now() - interval '15 days',
    now() - interval '15 days'
  ),
  (
    '00000000-0000-4000-8000-000000003007',
    'Redes de Computadores',
    'Turma com convite ativo e alunos de perfis diferentes.',
    2026,
    1,
    '00000000-0000-4000-8000-000000002004',
    now() - interval '14 days',
    now() - interval '14 days'
  ),
  (
    '00000000-0000-4000-8000-000000003008',
    'Governança de TI',
    'Disciplina híbrida para demonstrar mensagens direcionadas por campus.',
    2026,
    1,
    '00000000-0000-4000-8000-000000002005',
    now() - interval '13 days',
    now() - interval '13 days'
  ),
  (
    '00000000-0000-4000-8000-000000003009',
    'Segurança da Informação',
    'Disciplina planejada para o próximo semestre, ainda sem vínculos.',
    2026,
    2,
    '00000000-0000-4000-8000-000000002005',
    now() - interval '12 days',
    now() - interval '12 days'
  );

INSERT INTO students (
  id,
  student_id,
  name,
  phone,
  email,
  annotation,
  status,
  consent,
  created_at,
  updated_at
)
VALUES
  (
    '00000000-0000-4000-8000-000000004001',
    '2026001',
    'Ana Beatriz Lima',
    '5511991111111',
    'ana.lima@example.com',
    'Representante da turma.',
    'ACTIVE',
    true,
    now() - interval '15 days',
    now() - interval '12 days'
  ),
  (
    '00000000-0000-4000-8000-000000004002',
    '2026002',
    'Bruno Henrique Souza',
    '5511992222222',
    'bruno.souza@example.com',
    'Prefere comunicados por WhatsApp.',
    'ACTIVE',
    true,
    now() - interval '15 days',
    now() - interval '11 days'
  ),
  (
    '00000000-0000-4000-8000-000000004003',
    '2026003',
    'Carla Mendes Rocha',
    '5521993333333',
    'carla.rocha@example.com',
    'Aluno de outro campus cursando disciplina optativa.',
    'ACTIVE',
    true,
    now() - interval '14 days',
    now() - interval '10 days'
  ),
  (
    '00000000-0000-4000-8000-000000004004',
    '2026004',
    'Diego Martins Alves',
    null,
    null,
    'Pré-cadastrado por importação; ainda não concluiu auto-cadastro.',
    'PENDING',
    false,
    now() - interval '13 days',
    now() - interval '13 days'
  ),
  (
    '00000000-0000-4000-8000-000000004005',
    '2026005',
    'Eduarda Nunes Ferreira',
    '5531995555555',
    'eduarda.ferreira@example.com',
    'Status trancado para demonstrar gestão global do aluno.',
    'LOCKED',
    true,
    now() - interval '12 days',
    now() - interval '8 days'
  ),
  (
    '00000000-0000-4000-8000-000000004006',
    '2026006',
    'Felipe Costa Ribeiro',
    '5541996666666',
    'felipe.ribeiro@example.com',
    'Aluno graduado em disciplina anterior.',
    'GRADUATED',
    true,
    now() - interval '11 days',
    now() - interval '7 days'
  ),
  (
    '00000000-0000-4000-8000-000000004007',
    '2026007',
    'Gabriela Torres Almeida',
    '5551997777777',
    'gabriela.almeida@example.com',
    'Aluno cancelado mantido para histórico.',
    'CANCELED',
    true,
    now() - interval '10 days',
    now() - interval '6 days'
  ),
  (
    '00000000-0000-4000-8000-000000004008',
    '2026008',
    'Henrique Barros Pereira',
    null,
    'henrique.pereira@example.com',
    'Sem telefone para demonstrar falha de canal WhatsApp.',
    'ACTIVE',
    true,
    now() - interval '9 days',
    now() - interval '5 days'
  );

INSERT INTO enrollments (
  id,
  discipline_id,
  student_id,
  self_registration_completed_at,
  self_registration_count,
  created_at,
  updated_at
)
VALUES
  (
    '00000000-0000-4000-8000-000000005001',
    '00000000-0000-4000-8000-000000003001',
    '00000000-0000-4000-8000-000000004001',
    now() - interval '12 days',
    1,
    now() - interval '15 days',
    now() - interval '12 days'
  ),
  (
    '00000000-0000-4000-8000-000000005002',
    '00000000-0000-4000-8000-000000003001',
    '00000000-0000-4000-8000-000000004002',
    now() - interval '11 days',
    1,
    now() - interval '15 days',
    now() - interval '11 days'
  ),
  (
    '00000000-0000-4000-8000-000000005003',
    '00000000-0000-4000-8000-000000003001',
    '00000000-0000-4000-8000-000000004004',
    null,
    0,
    now() - interval '13 days',
    now() - interval '13 days'
  ),
  (
    '00000000-0000-4000-8000-000000005004',
    '00000000-0000-4000-8000-000000003001',
    '00000000-0000-4000-8000-000000004008',
    now() - interval '5 days',
    1,
    now() - interval '9 days',
    now() - interval '5 days'
  ),
  (
    '00000000-0000-4000-8000-000000005005',
    '00000000-0000-4000-8000-000000003002',
    '00000000-0000-4000-8000-000000004001',
    now() - interval '12 days',
    1,
    now() - interval '14 days',
    now() - interval '12 days'
  ),
  (
    '00000000-0000-4000-8000-000000005006',
    '00000000-0000-4000-8000-000000003002',
    '00000000-0000-4000-8000-000000004003',
    now() - interval '10 days',
    1,
    now() - interval '14 days',
    now() - interval '10 days'
  ),
  (
    '00000000-0000-4000-8000-000000005007',
    '00000000-0000-4000-8000-000000003003',
    '00000000-0000-4000-8000-000000004002',
    now() - interval '11 days',
    1,
    now() - interval '13 days',
    now() - interval '11 days'
  ),
  (
    '00000000-0000-4000-8000-000000005008',
    '00000000-0000-4000-8000-000000003003',
    '00000000-0000-4000-8000-000000004005',
    now() - interval '8 days',
    1,
    now() - interval '12 days',
    now() - interval '8 days'
  ),
  (
    '00000000-0000-4000-8000-000000005009',
    '00000000-0000-4000-8000-000000003004',
    '00000000-0000-4000-8000-000000004003',
    now() - interval '10 days',
    1,
    now() - interval '11 days',
    now() - interval '10 days'
  ),
  (
    '00000000-0000-4000-8000-000000005010',
    '00000000-0000-4000-8000-000000003004',
    '00000000-0000-4000-8000-000000004006',
    now() - interval '7 days',
    1,
    now() - interval '10 days',
    now() - interval '7 days'
  ),
  (
    '00000000-0000-4000-8000-000000005011',
    '00000000-0000-4000-8000-000000003004',
    '00000000-0000-4000-8000-000000004007',
    now() - interval '6 days',
    1,
    now() - interval '10 days',
    now() - interval '6 days'
  ),
  (
    '00000000-0000-4000-8000-000000005012',
    '00000000-0000-4000-8000-000000003006',
    '00000000-0000-4000-8000-000000004001',
    now() - interval '12 days',
    1,
    now() - interval '9 days',
    now() - interval '8 days'
  ),
  (
    '00000000-0000-4000-8000-000000005013',
    '00000000-0000-4000-8000-000000003006',
    '00000000-0000-4000-8000-000000004006',
    now() - interval '7 days',
    1,
    now() - interval '9 days',
    now() - interval '7 days'
  ),
  (
    '00000000-0000-4000-8000-000000005014',
    '00000000-0000-4000-8000-000000003007',
    '00000000-0000-4000-8000-000000004003',
    now() - interval '10 days',
    1,
    now() - interval '8 days',
    now() - interval '8 days'
  ),
  (
    '00000000-0000-4000-8000-000000005015',
    '00000000-0000-4000-8000-000000003007',
    '00000000-0000-4000-8000-000000004004',
    null,
    0,
    now() - interval '8 days',
    now() - interval '8 days'
  ),
  (
    '00000000-0000-4000-8000-000000005016',
    '00000000-0000-4000-8000-000000003008',
    '00000000-0000-4000-8000-000000004002',
    now() - interval '11 days',
    1,
    now() - interval '7 days',
    now() - interval '7 days'
  ),
  (
    '00000000-0000-4000-8000-000000005017',
    '00000000-0000-4000-8000-000000003008',
    '00000000-0000-4000-8000-000000004005',
    now() - interval '8 days',
    1,
    now() - interval '7 days',
    now() - interval '7 days'
  ),
  (
    '00000000-0000-4000-8000-000000005018',
    '00000000-0000-4000-8000-000000003008',
    '00000000-0000-4000-8000-000000004008',
    now() - interval '5 days',
    1,
    now() - interval '7 days',
    now() - interval '5 days'
  );

INSERT INTO invites (id, discipline_id, code, expires_at, active, created_at, updated_at)
VALUES
  (
    '00000000-0000-4000-8000-000000006001',
    '00000000-0000-4000-8000-000000003001',
    'TCC2026',
    now() + interval '14 days',
    true,
    now() - interval '12 days',
    now() - interval '12 days'
  ),
  (
    '00000000-0000-4000-8000-000000006002',
    '00000000-0000-4000-8000-000000003004',
    'WEB2026',
    now() + interval '7 days',
    true,
    now() - interval '8 days',
    now() - interval '8 days'
  ),
  (
    '00000000-0000-4000-8000-000000006003',
    '00000000-0000-4000-8000-000000003002',
    'BD2OLD',
    now() - interval '1 day',
    false,
    now() - interval '20 days',
    now() - interval '1 day'
  ),
  (
    '00000000-0000-4000-8000-000000006004',
    '00000000-0000-4000-8000-000000003007',
    'REDES26',
    now() + interval '21 days',
    true,
    now() - interval '7 days',
    now() - interval '7 days'
  ),
  (
    '00000000-0000-4000-8000-000000006005',
    '00000000-0000-4000-8000-000000003008',
    'GOVTI26',
    null,
    true,
    now() - interval '6 days',
    now() - interval '6 days'
  );

INSERT INTO message_logs (
  id,
  student_id,
  channel,
  success,
  error_text,
  subject,
  body,
  attachment_names,
  attachment_count,
  created_at
)
VALUES
  (
    '00000000-0000-4000-8000-000000007001',
    '00000000-0000-4000-8000-000000004001',
    'EMAIL',
    true,
    null,
    'Boas-vindas ao semestre',
    'Lembrete de apresentação do plano de ensino.',
    null,
    0,
    now() - interval '6 days'
  ),
  (
    '00000000-0000-4000-8000-000000007002',
    '00000000-0000-4000-8000-000000004002',
    'WHATSAPP',
    true,
    null,
    'Boas-vindas ao semestre',
    'Lembrete de apresentação do plano de ensino.',
    null,
    0,
    now() - interval '6 days'
  ),
  (
    '00000000-0000-4000-8000-000000007003',
    '00000000-0000-4000-8000-000000004008',
    'WHATSAPP',
    false,
    'Aluno sem telefone cadastrado.',
    'Boas-vindas ao semestre',
    'Lembrete de apresentação do plano de ensino.',
    null,
    0,
    now() - interval '6 days'
  );

COMMIT;
