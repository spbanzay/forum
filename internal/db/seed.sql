-- Пользователи
INSERT INTO users (email, username, password) VALUES
('alice@example.com', 'alice', '$2b$12$xJf01AYDPPABvVRQ2.NNhOOMzJl8FSrvJ9Vv0McQSH4.hqeX8KU4S'),
('bob@example.com', 'bob', '$2b$12$Cm4/O0HObp2AtiYeQ9FQl.KwfiqqutcK6uEzHe9Nz0uLhYz1AiunC');

-- Категории
INSERT INTO categories (name) VALUES
('General'), ('GoLang'), ('DevOps');

-- Посты
INSERT INTO posts (user_id, title, content) VALUES
(1, 'Добро пожаловать!', 'Это первый пост на форуме.'),
(2, 'Go — лучший язык', 'Обсудим плюсы и минусы Go.');

-- Связь постов с категориями
INSERT INTO post_categories (post_id, category_id) VALUES
(1, 1), (2, 2);

-- Комментарии
INSERT INTO comments (post_id, user_id, content) VALUES
(1, 2, 'Согласен!'),
(2, 1, 'Не уверен, но интересно.');

-- Лайки постов
INSERT INTO post_likes (post_id, user_id, is_like) VALUES
(1, 2, TRUE), (2, 1, FALSE);

INSERT INTO posts (user_id, title, content) VALUES (1, 'Футбол: Почему Лига чемпионов так популярна?', 'Лига чемпионов собирает лучшие клубы Европы, создавая неповторимую атмосферу соревнований. Каждый матч становится настоящим шоу, а болельщики ждут этих встреч с огромным нетерпением.');
INSERT INTO post_categories (post_id, category_id) VALUES (3, 1);
INSERT INTO comments (post_id, user_id, content) VALUES (3, 2, 'Полностью согласен, особенно нравится стадия плей-офф.');
INSERT INTO comments (post_id, user_id, content) VALUES (3, 1, 'Атмосфера на стадионах действительно потрясающая!');
INSERT INTO posts (user_id, title, content) VALUES (2, 'Бокс: Возвращение легенды на ринг', 'Недавно прошёл бой, где бывший чемпион мира вернулся в ринг после долгого перерыва. Он продемонстрировал отличную форму и напомнил фанатам, почему его считают одним из лучших.');
INSERT INTO post_categories (post_id, category_id) VALUES (4, 2);
INSERT INTO comments (post_id, user_id, content) VALUES (4, 1, 'Смотрел бой — очень достойное возвращение.');
INSERT INTO comments (post_id, user_id, content) VALUES (4, 2, 'Ждём следующего поединка!');
INSERT INTO posts (user_id, title, content) VALUES (1, 'Баскетбол: Новый сезон НБА стартовал', 'Сезон НБА начался с напряжённых матчей и неожиданностей. Молодые игроки показывают зрелую игру, а фавориты не всегда побеждают.');
INSERT INTO post_categories (post_id, category_id) VALUES (5, 3);
INSERT INTO comments (post_id, user_id, content) VALUES (5, 2, 'Интриги будет много в этом году.');