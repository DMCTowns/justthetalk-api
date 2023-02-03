-- This file is part of the JUSTtheTalkAPI distribution (https://github.com/jdudmesh/justthetalk-api).
-- Copyright (c) 2021 John Dudmesh.

-- This program is free software: you can redistribute it and/or modify
-- it under the terms of the GNU General Public License as published by
-- the Free Software Foundation, version 3.

-- This program is distributed in the hope that it will be useful, but
-- WITHOUT ANY WARRANTY; without even the implied warranty of
-- MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
-- General Public License for more details.

-- You should have received a copy of the GNU General Public License
-- along with this program. If not, see <http://www.gnu.org/licenses/>.


-- ssh -N -L 3307:beta.justthetalk.com:3306 jdudmesh@beta.justthetalk.com


-- create user 'notthetalk_app'@'%' identified WITH mysql_native_password by 'BeXuPHa3uC4c';
-- grant all privileges on notthetalk.* to 'notthetalk_app'@'%';
-- flush privileges;

-- INSERT INTO `notthetalk`.`user` (`version`, `account_expired`, `account_locked`, `bio`, `email`, `enabled`, `password`, `password_expired`, `username`, `display_email`, `created_date`, `last_updated`, `last_login_date`, `subject`, `email_verified`)
-- VALUES ('1', 0, 0, 'test user', 'test@notthetalk.com', 1, 'c775e7b757ede630cd0aa1113bd102661ab38829ca52a6422ab782862f268646', 0, 'testuser1', 1, '2011-02-27 23:58:48', '2014-01-30 22:36:13', '2013-03-16 07:32:38', 'johndudmesh123456', 1);

-- delete ud  from user_discussion ud inner join (select min(id) dupe_id from user_discussion group by user_id, discussion_id having count(*) > 1) ud2 on ud.id = ud2.dupe_id;
drop index idx_userdiscussion_userid_discussionid on user_discussion;
create unique index idx_userdiscussion_userid_discussionid on user_discussion(user_id, discussion_id);

-- create table discussion_activity (
--     discussion_id bigint not null unique references discussion(id),
--     post_count int,
--     last_created_date datetime
-- );

-- create index idx_discussion_activity_post_count on discussion_activity(post_count);

-- alter table ignore_user add column created_date datetime not null default UTC_TIMESTAMP();
-- alter table user_options add column view_type varchar(16) not null default 'latest';
-- alter table user_options add column subs_fetch_order int not null default 0;
-- alter table user_options add column subs_fetch_order int not null default 0;
-- alter table front_page_entry add column zorder int not null default 0;

-- create table user_post_bookmark (
--     id int not null primary key,
--     version int not null default 1,
--     created_date datetime not null default UTC_TIMESTAMP(),
--     user_id int not null references user(id),
--     post_id int not null references post(id)
-- );

-- alter table user add column email_verified int not null default 0;
-- update user set email_verified = 1 where id > 0;

-- alter table password_reset add column created_date datetime not null default UTC_TIMESTAMP();
-- alter table password_reset add column ip_address varchar(15);
drop index idx_password_reset_reset_key on password_reset;
create unique index idx_password_reset_reset_key on password_reset(reset_key);

-- alter table user_discussion add column created_date datetime not null default UTC_TIMESTAMP();

-- alter table front_page_entry  modify last_post datetime(6) null;
drop index idx_front_page_entry_last_post on front_page_entry;
create index idx_front_page_entry_last_post on front_page_entry(last_post);

---------------------------------------------

DROP PROCEDURE IF EXISTS get_folders;
DELIMITER //
CREATE PROCEDURE get_folders()
BEGIN

    select f.*,
    d.discussion_count
    from folder f
    inner join (select folder_id, count(*) discussion_count from discussion d group by folder_id) d
    on d.folder_id = f.id
    where f.type in (0, 3);

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_front_page_entries;
DELIMITER //
CREATE PROCEDURE get_front_page_entries(IN $page_size int)
BEGIN

    select d.id discussion_id,
    d.title discussion_name,
    f.id folder_id,
    f.folder_key,
    f.description folder_name,
    d.last_post,
    d.post_count,
    case when f.type = 0 then 0 else 1 end admin_only
    from discussion d
    inner join folder f
    on d.folder_id = f.id
    order by d.last_post desc
    limit $page_size;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_active_subscriptions;
DELIMITER //
CREATE PROCEDURE get_active_subscriptions(IN $user_id bigint, IN $page_start int, IN $page_size int)
BEGIN

    declare $is_admin int;

    select count(*) into $is_admin from user_role where user_id = $user_id and role_id in (2, 3);

    select d.id discussion_id,
    d.title discussion_name,
    f.id folder_id,
    f.folder_key,
    f.description folder_name,
    d.last_post,
    d.post_count,
    case when f.type = 0 then 0 else 1 end admin_only,
    ud.last_post_count last_post_read_count,
    ud.last_post last_post_read_date,
    ud.last_post_id last_post_read_id
    from discussion d
    inner join folder f
    on d.folder_id = f.id
    inner join subscription s
    on d.id = s.discussion_id
    left join user_discussion ud
    on s.discussion_id = ud.discussion_id
    and s.user_id = ud.user_id
    where s.user_id = $user_id
    and ((f.type = 0 and $is_admin >= 0) or (f.type = 3 and $is_admin > 0))
    order by d.last_post desc
    limit $page_start, $page_size;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_folder_discussions;
DELIMITER //
CREATE PROCEDURE get_folder_discussions(IN $folder_id bigint, IN $user_id bigint, IN $page_start int, IN $page_size int)
BEGIN

    select fp.discussion_id,
    fp.discussion_name,
    fp.folder_id,
    fp.folder_key,
    fp.folder_name,
    fp.last_post,
    fp.post_count,
    fp.last_post_id,
    coalesce(ud.last_post_count, 0) last_post_read_count,
    ud.last_post last_post_read_date,
    ud.last_post_id last_post_read_id
    from front_page_entry fp
    left join (select * from user_discussion where user_id = $user_id) ud
    on fp.discussion_id = ud.discussion_id
    where fp.folder_id = $folder_id
    order by fp.zorder desc, fp.last_post desc
    limit $page_start, $page_size;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_folder_discussions_before;
DELIMITER //
CREATE PROCEDURE get_folder_discussions_before(IN $folder_id bigint, IN $user_id bigint, IN $date_before datetime, IN $page_size int)
BEGIN

    select fp.discussion_id,
    fp.discussion_name,
    fp.folder_id,
    fp.folder_key,
    fp.folder_name,
    fp.last_post,
    fp.post_count,
    fp.last_post_id,
    coalesce(ud.last_post_count, 0) last_post_read_count,
    ud.last_post last_post_read_date,
    ud.last_post_id last_post_read_id
    from front_page_entry fp
    left join (select * from user_discussion where user_id = $user_id) ud
    on fp.discussion_id = ud.discussion_id
    where fp.folder_id = $folder_id
    and fp.last_post < $date_before
    order by fp.zorder desc, fp.last_post desc
    limit $page_size;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_discussion;
DELIMITER //
CREATE PROCEDURE get_discussion(IN $discussion_id bigint)
BEGIN

    select d.id,
    d.version,
    d.created_date,
    d.folder_id,
    d.header,
    d.last_post,
    d.title,
    d.user_id,
    u.username,
    case coalesce(d.deleted, 0) when 1 then 1 else 0 end deleted,
    case d.locked when 1 then 1 else 0 end locked,
    d.post_count,
    d.zorder,
    d.status,
    case d.premoderate when 1 then 1 else 0 end premoderate,
    d.last_updated,
    d.last_post_id
    from discussion d
    inner join folder f
    on d.folder_id = f.id
    inner join user u
    on d.user_id = u.id
    where d.id = $discussion_id;

END //
DELIMITER ;


DROP PROCEDURE IF EXISTS get_blocked_discussion_users;
DELIMITER //
CREATE PROCEDURE get_blocked_discussion_users(IN $discussion_id bigint)
BEGIN

    select id,
    version,
    discussion_id,
    user_id,
    case coalesce(user_status, 0) when 0 then 0 else 1 end user_status
    from discussion_user
    where discussion_id = $discussion_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS block_discussion_user;
DELIMITER //
CREATE PROCEDURE block_discussion_user(IN $discussion_id bigint, IN $user_id bigint, IN $state int)
BEGIN

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    start transaction;

    delete from discussion_user
    where discussion_id = $discussion_id
    and user_id = $user_id;

    if $state = 1 then

        insert into discussion_user (
        version,
        discussion_id,
        user_id,
        user_status)
        values (
            1,
            $discussion_id,
            $user_id,
            -1
        );

    end if;

    commit work;

    call get_blocked_discussion_users($discussion_id);

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_discussion_posts;
DELIMITER //
CREATE PROCEDURE get_discussion_posts(IN $user_id int, IN $folder_id bigint, IN $discussion_id bigint, IN $requested_start int, IN $page_size int)
BEGIN

	declare $page_start int;

    select $requested_start into $page_start;

    if coalesce($page_start, 0) = 0 then
		select last_post_count into $page_start
		from user_discussion
		where user_id = $user_id
		and discussion_id = $discussion_id;
	end if;

    if coalesce($page_start, 0) = 0 then
		select 1 into $page_start;
	end if;

    select p.id,
    p.version,
    p.created_date,
    p.discussion_id,
    d.status discussion_status,
    p.text,
    p.user_id,
    case p.deleted when 1 then 1 else 0 end deleted,
    p.moderation_result,
    p.moderation_score,
    p.status,
    p.last_edit_date,
    case p.markdown when 1 then 1 else 0 end markdown,
    p.post_count,
    p.post_num,
    u.id user_id,
    u.username,
    case u.enabled when 1 then 1 else 0 end user_enabled,
    case u.account_locked when 1 then 1 else 0 end user_locked,
    case u.account_expired when 1 then 1 else 0 end user_expired,
    case coalesce(o.watch, 0) when 1 then 1 else 0 end user_watch,
    case coalesce(o.premoderate) when 1 then 1 else 0 end user_premod
    from post p
    inner join discussion d
    on p.discussion_id = d.id
    inner join folder f
    on d.folder_id = f.id
    inner join user u
    on p.user_id = u.id
    left join user_options o
    on u.id = o.user_id
    where p.discussion_id = $discussion_id
    and d.folder_id = $folder_id
    and p.post_num >= $page_start
    and p.post_num < $page_start + $page_size
    order by p.post_num;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_post;
DELIMITER //
CREATE PROCEDURE get_post(IN $post_id bigint)
BEGIN

    select p.id,
    p.version,
    p.created_date,
    p.discussion_id,
    d.status discussion_status,
    p.text,
    p.user_id,
    case p.deleted when 1 then 1 else 0 end deleted,
    p.moderation_result,
    p.moderation_score,
    p.status,
    p.last_edit_date,
    case p.markdown when 1 then 1 else 0 end markdown,
    p.post_count,
    p.post_num,
    u.id user_id,
    u.username,
    case u.enabled when 1 then 1 else 0 end user_enabled,
    case u.account_locked when 1 then 1 else 0 end user_locked,
    case u.account_expired when 1 then 1 else 0 end user_expired,
    case coalesce(o.watch, 0) when 1 then 1 else 0 end user_watch,
    case coalesce(o.premoderate) when 1 then 1 else 0 end user_premod
    from post p
    inner join discussion d
    on p.discussion_id = d.id
    inner join folder f
    on d.folder_id = f.id
    inner join user u
    on p.user_id = u.id
    left join user_options o
    on u.id = o.user_id
    where p.id = $post_id;

END //
DELIMITER ;


DROP PROCEDURE IF EXISTS set_post_status;
DELIMITER //
CREATE PROCEDURE set_post_status(IN $discussion_id bigint, IN $post_id bigint, IN $status int, IN $moderation_result int)
BEGIN


    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    start transaction;

    update post
    set status = $status,
    moderation_result = $moderation_result
    where id = $post_id
    and discussion_id = $discussion_id;

    delete from moderation_queue
    where post_id = $post_id;

    commit work;

    call get_post($post_id);

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS find_user;
DELIMITER //
CREATE PROCEDURE find_user(IN $username varchar(24), IN $password varchar(64))
BEGIN

    select u.id
    from user u
    where (u.username = $username or u.email = $username)
    and u.password = $password;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS find_user_by_email;
DELIMITER //
CREATE PROCEDURE find_user_by_email(IN $email varchar(24))
BEGIN

    select u.id
    from user u
    where u.email = $email;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS check_user;
DELIMITER //
CREATE PROCEDURE check_user(IN $username varchar(24), IN $email varchar(24))
BEGIN

    select count(*) count_of_existing
    from user u
    where u.username = $username or u.email = $email;

END //
DELIMITER ;


DROP PROCEDURE IF EXISTS get_user;
DELIMITER //
CREATE PROCEDURE get_user(IN $user_id bigint)
BEGIN

    declare $is_admin int;

    select count(*) into $is_admin from user_role where user_id = $user_id and role_id in (2, 3);

    select u.id,
    u.version,
    u.bio,
    u.email,
    u.password,
    u.username,
    u.created_date,
    u.last_updated,
    u.last_login_date,
    case u.account_expired when 1 then 1 else 0 end account_expired,
    case u.account_locked when 1 then 1 else 0 end account_locked,
    case u.enabled when 1 then 1 else 0 end enabled,
    case u.password_expired when 1 then 1 else 0 end password_expired,
    case u.display_email when 1 then 1 else 0 end display_email,
    case when $is_admin > 0 then 1 else 0 end is_admin,
    case o.premoderate when 1 then 1 else 0 end is_premoderate,
    case o.watch when 1 then 1 else 0 end is_watch,
    case o.sort_folders_by_activity when 1 then 1 else 0 end sort_folders_by_activity,
    case o.auto_subs when 1 then 1 else 0 end auto_subs,
    u.email_verified,
    o.view_type,
    o.subs_fetch_order
    from user u
    left join user_options o
    on u.id = o.user_id
    where u.id = $user_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS create_user;
DELIMITER //
CREATE PROCEDURE create_user(IN $email varchar(64), IN $username varchar(24), IN $password varchar(255))
BEGIN

    declare $last_user_id int;
    declare $current_timestamp datetime;

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    start transaction;


    select UTC_TIMESTAMP() into $current_timestamp;

    INSERT INTO user (
    version,
    account_expired,
    account_locked,
    bio,
    email,
    enabled,
    password,
    password_expired,
    username,
    display_email,
    created_date,
    last_updated,
    last_login_date,
    email_verified)
    VALUES (
    1,
    0,
    1,
    '',
    $email,
    1,
    $password,
    0,
    $username,
    0,
    $current_timestamp,
    $current_timestamp,
    $current_timestamp,
    0);

    set $last_user_id = LAST_INSERT_ID();

    INSERT INTO user_options (
    version,
    auto_subs,
    user_id,
    sort_folders_by_activity,
    markdown,
    premoderate,
    subscription_sort_order,
    watch,
    view_type)
    VALUES (
    1,
    1,
    $last_user_id,
    0,
    0,
    0,
    0,
    1,
    'latest');

    commit work;

    call get_user($last_user_id);

END //
DELIMITER ;



DROP PROCEDURE IF EXISTS create_signup_confirmation;
DELIMITER //
CREATE PROCEDURE create_signup_confirmation(IN $user_id bigint, IN $confirmation_key varchar(128))
BEGIN

    declare $last_user_id int;

    INSERT INTO signup_confirmation (
    version,
    confirmation_key,
    created_date,
    ip_address,
    last_updated,
    user_id)
    VALUES (
    1,
    $confirmation_key,
    UTC_TIMESTAMP(),
    null,
    UTC_TIMESTAMP(),
    $user_id);

    set $last_user_id = LAST_INSERT_ID();

    select s.*, u.username
    from signup_confirmation s
    inner join user u
    on s.user_id = u.id
    where s.id = $last_user_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_user_folder_bookmarks;
DELIMITER //
CREATE PROCEDURE get_user_folder_bookmarks(IN $user_id bigint)
BEGIN

    select *
    from folder_subscription
    where user_id = $user_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS update_user_password;
DELIMITER //
CREATE PROCEDURE update_user_password(IN $user_id bigint, IN $password_hash varchar(64))
BEGIN

    update user
    set password = $password_hash,
    password_expired = 0
    where id = $user_id;

    call get_user($user_id);

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS update_user_bio;
DELIMITER //
CREATE PROCEDURE update_user_bio(IN $user_id bigint, IN $bio varchar(1024))
BEGIN

    update user
    set bio = $bio
    where id = $user_id;

    call get_user($user_id);

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS update_user_autosubscribe;
DELIMITER //
CREATE PROCEDURE update_user_autosubscribe(IN $user_id bigint, IN $state int)
BEGIN

    update user_options
    set auto_subs = $state
    where user_id = $user_id;

    call get_user($user_id);

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS update_user_subsfetchorder;
DELIMITER //
CREATE PROCEDURE update_user_subsfetchorder(IN $user_id bigint, IN $state int)
BEGIN

    update user_options
    set subs_fetch_order  = $state
    where user_id = $user_id;

    call get_user($user_id);

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS update_user_foldersort;
DELIMITER //
CREATE PROCEDURE update_user_foldersort(IN $user_id bigint, IN $state int)
BEGIN

    update user_options
    set sort_folders_by_activity = $state
    where user_id = $user_id;

    call get_user($user_id);

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS update_user_viewtype;
DELIMITER //
CREATE PROCEDURE update_user_viewtype(IN $user_id bigint, IN $view_type varchar(16))
BEGIN

    update user_options
    set view_type = $view_type
    where user_id = $user_id;

    call get_user($user_id);

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_discussion_subscription_status;
DELIMITER //
CREATE PROCEDURE get_discussion_subscription_status(IN $user_id bigint, IN $discussion_id bigint)
BEGIN

    declare $is_subscribed bigint;

    select count(*) into $is_subscribed
    from subscription s
    where s.user_id = $user_id
    and s.discussion_id = $discussion_id;

    if $is_subscribed = 0 then

        select count(*) into $is_subscribed
        from folder_subscription fs
        inner join discussion d
        on fs.folder_id = d.folder_id
        where fs.user_id = $user_id
        and d.id = $discussion_id
        and not fs.id in (select subscription_id from folder_subscription_exception where discussion_id = $discussion_id);

    end if;

    select $is_subscribed is_subscribed;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_folder_subscription_status;
DELIMITER //
CREATE PROCEDURE get_folder_subscription_status(IN $user_id bigint, IN $folder_id bigint)
BEGIN

    declare $is_subscribed bigint;

    select count(*) into $is_subscribed
    from folder_subscription fs
    where fs.user_id = $user_id
    and fs.folder_id = $folder_id;

    select $is_subscribed is_subscribed;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS update_user_discussion_subscription;
DELIMITER //
CREATE PROCEDURE update_user_discussion_subscription(IN $user_id bigint, IN $discussion_id bigint, IN $state int)
BEGIN

    declare $folder_subscription_id bigint;

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    start transaction;

    select s.id into $folder_subscription_id
    from folder_subscription s
    inner join discussion d
    on s.folder_id = d.folder_id
    where s.user_id = $user_id
    and d.id = $discussion_id;

    if isnull($folder_subscription_id) then

        delete
        from subscription
        where user_id = $user_id
        and discussion_id = $discussion_id;

        if $state <> 0 then

            INSERT INTO subscription (
            version,
            discussion_id,
            user_id)
            VALUES (
            1,
            $discussion_id,
            $user_id);

        end if;

    else

        delete from folder_subscription_exception
        where subscription_id = $folder_subscription_id;

        if $state = 0 then
            insert into folder_subscription_exception(version, discussion_id, subscription_id) values(1, $discussion_id, $folder_subscription_id);
        end if;

    end if;

    commit work;

    call get_user_discussion_subscriptions($user_id);

END //
DELIMITER ;


DROP PROCEDURE IF EXISTS update_user_folder_subscription;
DELIMITER //
CREATE PROCEDURE update_user_folder_subscription(IN $user_id bigint, IN $folder_id bigint, IN $state int)
BEGIN

    declare $subscription_id bigint;

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    start transaction;

    select id into $subscription_id
    from folder_subscription
    where user_id = $user_id
    and folder_id = $folder_id;

    delete from folder_subscription_exception where subscription_id = $subscription_id;
    delete from folder_subscription where id = $subscription_id;

    if $state <> 0 then

        INSERT INTO folder_subscription (
        version,
        created_date,
        folder_id,
        last_read,
        user_id)
        VALUES (
        1,
        UTC_TIMESTAMP(),
        $folder_id,
        UTC_TIMESTAMP(),
        $user_id);

    end if;

    commit work;

    call get_user_folder_subscriptions($user_id);

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS update_user_folder_subscription_exception;
DELIMITER //
CREATE PROCEDURE update_user_folder_subscription_exception(IN $user_id bigint, IN $discussion_id bigint, IN $state int)
BEGIN

    declare $folder_id bigint;
    declare $subscription_id bigint;

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    start transaction;

    select folder_id into $folder_id
    from discussion d
    where id = $discussion_id;

    select id into $subscription_id
    from folder_subscription
    where user_id = $user_id
    and folder_id = $folder_id;

    delete from folder_subscription_exception
    where subscription_id = $subscription_id
    and discussion_id = $discussion_id;

    if $state <> 0 then

        INSERT INTO folder_subscription_exception (
        version,
        discussion_id,
        subscription_id)
        VALUES (
        1,
        $discussion_id,
        $subscription_id);

    end if;

    commit work;

    call get_user_folder_subscription_exceptions($user_id);

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_user_folder_subscriptions;
DELIMITER //
CREATE PROCEDURE get_user_folder_subscriptions(IN $user_id bigint)
BEGIN

    select *
    from folder_subscription
    where user_id = $user_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_user_folder_subscription_exceptions;
DELIMITER //
CREATE PROCEDURE get_user_folder_subscription_exceptions(IN $user_id bigint)
BEGIN

    select s.*,
    e.discussion_id
    from folder_subscription s
    inner join folder_subscription_exception e
    on s.id = e.subscription_id
    where user_id = $user_id;

END //
DELIMITER ;


DROP PROCEDURE IF EXISTS get_user_discussion_subscriptions;
DELIMITER //
CREATE PROCEDURE get_user_discussion_subscriptions(IN $user_id bigint)
BEGIN

    select *
    from subscription
    where user_id = $user_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_user_discussion_bookmarks;
DELIMITER //
CREATE PROCEDURE get_user_discussion_bookmarks(IN $user_id bigint)
BEGIN

    select *
    from user_discussion
    where user_id = $user_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_user_discussion_bookmark;
DELIMITER //
CREATE PROCEDURE get_user_discussion_bookmark(IN $user_id bigint, IN $discussion_id bigint)
BEGIN

    select *
    from user_discussion
    where user_id = $user_id
    and discussion_id = $discussion_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS update_user_bookmark;
DELIMITER //
CREATE PROCEDURE update_user_bookmark(IN $user_id bigint, IN $discussion_id bigint, IN $last_post_id bigint, IN $last_post_count bigint, IN $last_post_date datetime)
BEGIN

    declare $bookmark_id bigint;
    declare $last_post_count_prev bigint;

    select id, last_post_count into $bookmark_id, $last_post_count_prev
    from user_discussion
    where user_id = $user_id
    and discussion_id = $discussion_id;

    if $bookmark_id is null then

        insert into user_discussion (
        created_date,
        version,
        user_id,
        discussion_id,
        last_updated,
        last_post,
        last_post_count,
        last_post_id)
        values (
        now(),
        0,
        $user_id,
        $discussion_id,
        UTC_TIMESTAMP(),
        $last_post_date,
        $last_post_count,
        $last_post_id);


    else

        if $last_post_count > $last_post_count_prev then

            update user_discussion
            set version = version + 1,
            last_updated = UTC_TIMESTAMP(),
            last_post = $last_post_date,
            last_post_count = $last_post_count,
            last_post_id = $last_post_id
            where id = $bookmark_id;

        end if;

    end if;

    select ud.*
    from user_discussion ud
    where ud.user_id = $user_id
    and ud.discussion_id = $discussion_id;


END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_frontpage_entry;
DELIMITER //
CREATE PROCEDURE get_frontpage_entry(IN $discussion_id int)
BEGIN

    select fp.discussion_id,
    fp.discussion_name,
    fp.folder_id,
    fp.folder_key,
    fp.folder_name,
    fp.last_post,
    fp.post_count,
    fp.last_post_id,
    null last_post_read_count,
    null last_post_read_date,
    null last_post_read_id
    from front_page_entry fp
    where fp.discussion_id = $discussion_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_frontpage_latest;
DELIMITER //
CREATE PROCEDURE get_frontpage_latest(IN $user_id int, IN $is_admin int, IN $page_start int, IN $page_size int)
BEGIN

    select fp.discussion_id,
    fp.discussion_name,
    fp.folder_id,
    fp.folder_key,
    fp.folder_name,
    fp.last_post,
    fp.post_count,
    fp.last_post_id,
    coalesce(ud.last_post_count, 0) last_post_read_count,
    ud.last_post last_post_read_date,
    ud.last_post_id last_post_read_id
    from front_page_entry fp
    left join (select * from user_discussion where user_id = $user_id) ud
    on fp.discussion_id = ud.discussion_id
    where ((fp.admin_only = 0) or (fp.admin_only and $is_admin > 0))
    order by fp.zorder desc, fp.last_post desc
    limit $page_start, $page_size;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_frontpage_subscriptions;
DELIMITER //
CREATE PROCEDURE get_frontpage_subscriptions(IN $user_id int, IN $is_admin int, IN $page_start int, IN $page_size int)
BEGIN

    declare $subs_fetch_order int;

    select subs_fetch_order into $subs_fetch_order from user_options where user_id = $user_id;

    if $subs_fetch_order = 0 then

        select fp.discussion_id,
        fp.discussion_name,
        fp.folder_id,
        fp.folder_key,
        fp.folder_name,
        fp.last_post,
        fp.post_count,
        fp.last_post_id,
        coalesce(ud.last_post_count, 0) last_post_read_count,
        ud.last_post last_post_read_date,
        ud.last_post_id last_post_read_id,
		case when (fp.post_count - coalesce(ud.last_post_count, 0)) > 0 then 1 else 0 end has_unread,
		(fp.post_count - coalesce(ud.last_post_count, 0)) unread_count
        from front_page_entry fp
        left join (select * from user_discussion where user_id = $user_id) ud
        on fp.discussion_id = ud.discussion_id
		inner join (select * from subscription where user_id = $user_id) subs
		on fp.discussion_id = subs.discussion_id
        where ((fp.admin_only = 0) or (fp.admin_only and $is_admin > 0))

		union

        select fp.discussion_id,
        fp.discussion_name,
        fp.folder_id,
        fp.folder_key,
        fp.folder_name,
        fp.last_post,
        fp.post_count,
        fp.last_post_id,
        coalesce(ud.last_post_count, 0) last_post_read_count,
        ud.last_post last_post_read_date,
        ud.last_post_id last_post_read_id,
		case when (fp.post_count - coalesce(ud.last_post_count, 0)) > 0 then 1 else 0 end has_unread,
		(fp.post_count - coalesce(ud.last_post_count, 0)) unread_count
        from front_page_entry fp
        left join (select * from user_discussion where user_id = $user_id) ud
        on fp.discussion_id = ud.discussion_id
		inner join (select * from folder_subscription where user_id = $user_id) subs
		on fp.folder_id = subs.folder_id
		and fp.folder_id not in (select discussion_id from folder_subscription_exception where subscription_id in (select id from folder_subscription where user_id = $user_id))
		and fp.last_post > subs.last_read
        and (fp.post_count - coalesce(ud.last_post_count, 0)) > 0
        where ((fp.admin_only = 0) or (fp.admin_only and $is_admin > 0))

		order by has_unread desc,
		last_post asc
		limit $page_start, $page_size;

    elseif $subs_fetch_order = 1 then

        select fp.discussion_id,
        fp.discussion_name,
        fp.folder_id,
        fp.folder_key,
        fp.folder_name,
        fp.last_post,
        fp.post_count,
        fp.last_post_id,
        coalesce(ud.last_post_count, 0) last_post_read_count,
        ud.last_post last_post_read_date,
        ud.last_post_id last_post_read_id,
		case when (fp.post_count - coalesce(ud.last_post_count, 0)) > 0 then 1 else 0 end has_unread,
		(fp.post_count - coalesce(ud.last_post_count, 0)) unread_count
        from front_page_entry fp
        left join (select * from user_discussion where user_id = $user_id) ud
        on fp.discussion_id = ud.discussion_id
		inner join (select * from subscription where user_id = $user_id) subs
		on fp.discussion_id = subs.discussion_id
        where ((fp.admin_only = 0) or (fp.admin_only and $is_admin > 0))

		union

        select fp.discussion_id,
        fp.discussion_name,
        fp.folder_id,
        fp.folder_key,
        fp.folder_name,
        fp.last_post,
        fp.post_count,
        fp.last_post_id,
        coalesce(ud.last_post_count, 0) last_post_read_count,
        ud.last_post last_post_read_date,
        ud.last_post_id last_post_read_id,
		case when (fp.post_count - coalesce(ud.last_post_count, 0)) > 0 then 1 else 0 end has_unread,
		(fp.post_count - coalesce(ud.last_post_count, 0)) unread_count
        from front_page_entry fp
        left join (select * from user_discussion where user_id = $user_id) ud
        on fp.discussion_id = ud.discussion_id
		inner join (select * from folder_subscription where user_id = $user_id) subs
		on fp.folder_id = subs.folder_id
		and fp.folder_id not in (select discussion_id from folder_subscription_exception where subscription_id in (select id from folder_subscription where user_id = $user_id))
		and fp.last_post > subs.last_read
        and (fp.post_count - coalesce(ud.last_post_count, 0)) > 0
        where ((fp.admin_only = 0) or (fp.admin_only and $is_admin > 0))

		order by has_unread desc,
		last_post desc
		limit $page_start, $page_size;

    elseif $subs_fetch_order = 2 then

        select fp.discussion_id,
        fp.discussion_name,
        fp.folder_id,
        fp.folder_key,
        fp.folder_name,
        fp.last_post,
        fp.post_count,
        fp.last_post_id,
        coalesce(ud.last_post_count, 0) last_post_read_count,
        ud.last_post last_post_read_date,
        ud.last_post_id last_post_read_id,
		case when (fp.post_count - coalesce(ud.last_post_count, 0)) > 0 then 1 else 0 end has_unread,
		(fp.post_count - coalesce(ud.last_post_count, 0)) unread_count
        from front_page_entry fp
        left join (select * from user_discussion where user_id = $user_id) ud
        on fp.discussion_id = ud.discussion_id
		inner join (select * from subscription where user_id = $user_id) subs
		on fp.discussion_id = subs.discussion_id
        where ((fp.admin_only = 0) or (fp.admin_only and $is_admin > 0))

		union

        select fp.discussion_id,
        fp.discussion_name,
        fp.folder_id,
        fp.folder_key,
        fp.folder_name,
        fp.last_post,
        fp.post_count,
        fp.last_post_id,
        coalesce(ud.last_post_count, 0) last_post_read_count,
        ud.last_post last_post_read_date,
        ud.last_post_id last_post_read_id,
		case when (fp.post_count - coalesce(ud.last_post_count, 0)) > 0 then 1 else 0 end has_unread,
		(fp.post_count - coalesce(ud.last_post_count, 0)) unread_count
        from front_page_entry fp
        left join (select * from user_discussion where user_id = $user_id) ud
        on fp.discussion_id = ud.discussion_id
		inner join (select * from folder_subscription where user_id = $user_id) subs
		on fp.folder_id = subs.folder_id
		and fp.folder_id not in (select discussion_id from folder_subscription_exception where subscription_id in (select id from folder_subscription where user_id = $user_id))
		and fp.last_post > subs.last_read
        and (fp.post_count - coalesce(ud.last_post_count, 0)) > 0
        where ((fp.admin_only = 0) or (fp.admin_only and $is_admin > 0))

		order by unread_count desc,
		last_post desc
		limit $page_start, $page_size;

    end if;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_frontpage_startedbyme;
DELIMITER //
CREATE PROCEDURE get_frontpage_startedbyme(IN $user_id int, IN $is_admin int, IN $page_start int, IN $page_size int)
BEGIN

    select fp.discussion_id,
    fp.discussion_name,
    fp.folder_id,
    fp.folder_key,
    fp.folder_name,
    fp.last_post,
    fp.post_count,
    fp.last_post_id,
    coalesce(ud.last_post_count, 0) last_post_read_count,
    ud.last_post last_post_read_date,
    ud.last_post_id last_post_read_id,
    case when (fp.post_count - coalesce(ud.last_post_count, 0)) > 0 then 1 else 0 end has_unread,
    (fp.post_count - coalesce(ud.last_post_count, 0)) unread_count
    from front_page_entry fp
    left join (select * from user_discussion where user_id = $user_id) ud
    on fp.discussion_id = ud.discussion_id
    inner join discussion d
    on fp.discussion_id = d.id
    where ((fp.admin_only = 0) or (fp.admin_only and $is_admin > 0))
    and d.user_id = $user_id
    order by fp.zorder desc,
    fp.last_post desc
    limit $page_start, $page_size;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_frontpage_mostactive;
DELIMITER //
CREATE PROCEDURE get_frontpage_mostactive(IN $user_id int, IN $is_admin int, IN $page_start int, IN $page_size int)
BEGIN

    select fp.discussion_id,
    fp.discussion_name,
    fp.folder_id,
    fp.folder_key,
    fp.folder_name,
    fp.last_post,
    fp.post_count,
    fp.last_post_id,
    coalesce(ud.last_post_count, 0) last_post_read_count,
    ud.last_post last_post_read_date,
    ud.last_post_id last_post_read_id,
    case when (fp.post_count - coalesce(ud.last_post_count, 0)) > 0 then 1 else 0 end has_unread,
    (fp.post_count - coalesce(ud.last_post_count, 0)) unread_count
    from front_page_entry fp
    left join (select * from user_discussion where user_id = $user_id) ud
    on fp.discussion_id = ud.discussion_id
    inner join discussion_activity da
    on fp.discussion_id = da.discussion_id
    where ((fp.admin_only = 0) or (fp.admin_only and $is_admin > 0))
    order by fp.zorder desc,
    da.post_count desc
    limit $page_start, $page_size;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_frontpage_latest_since;
DELIMITER //
CREATE PROCEDURE get_frontpage_latest_since(IN $user_id int, IN $is_admin int, IN $date_since datetime, IN $page_size int)
BEGIN

    select fp.discussion_id,
    fp.discussion_name,
    fp.folder_id,
    fp.folder_key,
    fp.folder_name,
    fp.last_post,
    fp.post_count,
    fp.last_post_id,
    coalesce(ud.last_post_count, 0) last_post_read_count,
    ud.last_post last_post_read_date,
    ud.last_post_id last_post_read_id
    from front_page_entry fp
    left join (select * from user_discussion where user_id = $user_id) ud
    on fp.discussion_id = ud.discussion_id
    where ((fp.admin_only = 0) or (fp.admin_only and $is_admin > 0))
    and fp.last_post > $date_since
    order by fp.zorder desc, fp.last_post desc
    limit $page_size;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_frontpage_mostactive_since;
DELIMITER //
CREATE PROCEDURE get_frontpage_mostactive_since(IN $user_id int, IN $is_admin int, IN $date_since datetime, IN $page_size int)
BEGIN

    select fp.discussion_id,
    fp.discussion_name,
    fp.folder_id,
    fp.folder_key,
    fp.folder_name,
    fp.last_post,
    fp.post_count,
    fp.last_post_id,
    coalesce(ud.last_post_count, 0) last_post_read_count,
    ud.last_post last_post_read_date,
    ud.last_post_id last_post_read_id,
    case when (fp.post_count - coalesce(ud.last_post_count, 0)) > 0 then 1 else 0 end has_unread,
    (fp.post_count - coalesce(ud.last_post_count, 0)) unread_count
    from front_page_entry fp
    left join (select * from user_discussion where user_id = $user_id) ud
    on fp.discussion_id = ud.discussion_id
    inner join discussion_activity da
    on fp.discussion_id = da.discussion_id
    where ((fp.admin_only = 0) or (fp.admin_only and $is_admin > 0))
    and fp.last_post > $date_since
    order by fp.zorder desc,
    da.post_count desc
    limit $page_size;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_frontpage_startedbyme_since;
DELIMITER //
CREATE PROCEDURE get_frontpage_startedbyme_since(IN $user_id int, IN $is_admin int, IN $date_since datetime, IN $page_size int)
BEGIN

    select fp.discussion_id,
    fp.discussion_name,
    fp.folder_id,
    fp.folder_key,
    fp.folder_name,
    fp.last_post,
    fp.post_count,
    fp.last_post_id,
    coalesce(ud.last_post_count, 0) last_post_read_count,
    ud.last_post last_post_read_date,
    ud.last_post_id last_post_read_id,
    case when (fp.post_count - coalesce(ud.last_post_count, 0)) > 0 then 1 else 0 end has_unread,
    (fp.post_count - coalesce(ud.last_post_count, 0)) unread_count
    from front_page_entry fp
    left join (select * from user_discussion where user_id = $user_id) ud
    on fp.discussion_id = ud.discussion_id
    inner join discussion d
    on fp.discussion_id = d.id
    where ((fp.admin_only = 0) or (fp.admin_only and $is_admin > 0))
    and d.user_id = $user_id
    and fp.last_post > $date_since
    order by fp.zorder desc,
    fp.last_post desc
    limit $page_size;

END //
DELIMITER ;

--

DROP PROCEDURE IF EXISTS get_frontpage_latest_before;
DELIMITER //
CREATE PROCEDURE get_frontpage_latest_before(IN $user_id int, IN $is_admin int, IN $date_before datetime, IN $page_size int)
BEGIN

    select fp.discussion_id,
    fp.discussion_name,
    fp.folder_id,
    fp.folder_key,
    fp.folder_name,
    fp.last_post,
    fp.post_count,
    fp.last_post_id,
    coalesce(ud.last_post_count, 0) last_post_read_count,
    ud.last_post last_post_read_date,
    ud.last_post_id last_post_read_id
    from front_page_entry fp
    left join (select * from user_discussion where user_id = $user_id) ud
    on fp.discussion_id = ud.discussion_id
    where ((fp.admin_only = 0) or (fp.admin_only and $is_admin > 0))
    and fp.last_post < $date_before
    order by fp.zorder desc, fp.last_post desc
    limit $page_size;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_frontpage_mostactive_before;
DELIMITER //
CREATE PROCEDURE get_frontpage_mostactive_before(IN $user_id int, IN $is_admin int, IN $date_before datetime, IN $page_size int)
BEGIN

    select fp.discussion_id,
    fp.discussion_name,
    fp.folder_id,
    fp.folder_key,
    fp.folder_name,
    fp.last_post,
    fp.post_count,
    fp.last_post_id,
    coalesce(ud.last_post_count, 0) last_post_read_count,
    ud.last_post last_post_read_date,
    ud.last_post_id last_post_read_id,
    case when (fp.post_count - coalesce(ud.last_post_count, 0)) > 0 then 1 else 0 end has_unread,
    (fp.post_count - coalesce(ud.last_post_count, 0)) unread_count
    from front_page_entry fp
    left join (select * from user_discussion where user_id = $user_id) ud
    on fp.discussion_id = ud.discussion_id
    inner join discussion_activity da
    on fp.discussion_id = da.discussion_id
    where ((fp.admin_only = 0) or (fp.admin_only and $is_admin > 0))
    and fp.last_post < $date_before
    order by fp.zorder desc,
    da.post_count desc
    limit $page_size;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_frontpage_startedbyme_before;
DELIMITER //
CREATE PROCEDURE get_frontpage_startedbyme_before(IN $user_id int, IN $is_admin int, IN $date_before datetime, IN $page_size int)
BEGIN

    select fp.discussion_id,
    fp.discussion_name,
    fp.folder_id,
    fp.folder_key,
    fp.folder_name,
    fp.last_post,
    fp.post_count,
    fp.last_post_id,
    coalesce(ud.last_post_count, 0) last_post_read_count,
    ud.last_post last_post_read_date,
    ud.last_post_id last_post_read_id,
    case when (fp.post_count - coalesce(ud.last_post_count, 0)) > 0 then 1 else 0 end has_unread,
    (fp.post_count - coalesce(ud.last_post_count, 0)) unread_count
    from front_page_entry fp
    left join (select * from user_discussion where user_id = $user_id) ud
    on fp.discussion_id = ud.discussion_id
    inner join discussion d
    on fp.discussion_id = d.id
    where ((fp.admin_only = 0) or (fp.admin_only and $is_admin > 0))
    and d.user_id = $user_id
    and fp.last_post < $date_before
    order by fp.zorder desc,
    fp.last_post desc
    limit $page_size;

END //
DELIMITER ;


DROP PROCEDURE IF EXISTS calculate_frontpage_mostactive;
DELIMITER //
CREATE PROCEDURE calculate_frontpage_mostactive()
BEGIN

    declare $last_post_date datetime;

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    select max(created_date) into $last_post_date from post;

	drop table if exists most_active;

    create temporary table most_active (
        discussion_id bigint not null references discussion(id),
        count_of_posts int,
        last_created_date datetime
    );

    insert into most_active(discussion_id, count_of_posts, last_created_date)
    select discussion_id, count(*) * 100000 count_of_posts, max(created_date) last_created_date
    from post
    where post.created_date > date_sub($last_post_date, interval 1 hour)
    group by discussion_id
    union all
    select discussion_id, count(*) * 10000 count_of_posts, max(created_date) last_created_date
    from post
    where post.created_date <= date_sub($last_post_date, interval 1 hour)
    and post.created_date > date_sub($last_post_date, interval 6 hour)
    group by discussion_id
    union all
    select discussion_id, count(*) * 100 count_of_posts, max(created_date) last_created_date
    from post
    where post.created_date <= date_sub($last_post_date, interval 6 hour)
    and post.created_date > date_sub($last_post_date, interval 24 hour)
    group by discussion_id
    union all
    select discussion_id, count(*) * 1 count_of_posts, max(created_date) last_created_date
    from post
    where post.created_date <= date_sub($last_post_date, interval 24 hour)
    and post.created_date > date_sub($last_post_date, interval 30 day)
    group by discussion_id;

    start transaction;

    delete from discussion_activity where discussion_id > 0;

    insert into discussion_activity(discussion_id, post_count, last_created_date)
    select discussion_id, sum(count_of_posts), max(last_created_date)
    from most_active
    group by discussion_id;

    update folder f set activity = 0 where id > 0;

    update folder f,
    (select d.folder_id, sum(count_of_posts) activity
    from most_active m
    inner join discussion d
    on m.discussion_id = d.id
    group by d.folder_id) a
    set f.activity = a.activity
    where f.id = a.folder_id;

    commit work;

    drop table most_active;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_user_blocked_discussion;
DELIMITER //
CREATE PROCEDURE get_user_blocked_discussion(IN $user_id int)
BEGIN

	select *
    from discussion_user
    where user_id = $user_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_user_ignored_users;
DELIMITER //
CREATE PROCEDURE get_user_ignored_users(IN $user_id int)
BEGIN

	select i.*, u.username ignored_username
    from ignore_user i
    inner join user u
    on i.ignored_user_id = u.id
    where user_id = $user_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS delete_user_ignored_users;
DELIMITER //
CREATE PROCEDURE delete_user_ignored_users(IN $user_id int, IN $ignored_user_id int)
BEGIN

    delete from ignore_user where user_id = $user_id and ignored_user_id = $ignored_user_id;

	select i.*, u.username ignored_username
    from ignore_user i
    inner join user u
    on i.ignored_user_id = u.id
    where user_id = $user_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS create_discussion_post;
DELIMITER //
CREATE PROCEDURE create_discussion_post(IN $folder_id bigint, IN $discussion_id bigint, IN $body_text varchar(8192), IN $post_status int, IN $user_id int)
BEGIN

    declare $post_num int;
    declare $last_post_id int;
    declare $current_timestamp datetime(6);

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    select UTC_TIMESTAMP(6) into $current_timestamp;

    start transaction;

    select count(*) + 1 into $post_num from post where discussion_id = $discussion_id;

    INSERT INTO post (
        version,
        created_date,
        discussion_id,
        text,
        user_id,
        deleted,
        moderation_result,
        moderation_score,
        status,
        last_edit_date,
        markdown,
        post_count,
        post_num)
    VALUES (
        1,
        $current_timestamp,
        $discussion_id,
        $body_text,
        $user_id,
        0,
        0,
        0,
        $post_status,
        null,
        0,
        null,
        $post_num);

	set $last_post_id = LAST_INSERT_ID();

    update discussion
    set post_count = $post_num,
    last_post = $current_timestamp,
    last_post_id = $last_post_id
    where id = $discussion_id;

    if $post_status = 1 or $post_status = 4 then
        INSERT INTO moderation_queue (
        version,
        created_date,
        post_id)
        VALUES (
        1,
        $current_timestamp,
        $last_post_id);
    end if;

    if $post_status = 0 or $post_status = 3 or $post_status = 4 then

        delete from front_page_entry where discussion_id = $discussion_id;

        insert into front_page_entry (version, discussion_id, discussion_name, zorder, folder_id, folder_key, folder_name, last_post, last_post_id, post_count, admin_only)
        select 0, d.id, d.title, d.zorder, d.folder_id, f.folder_key, f.description, $current_timestamp, $last_post_id, $post_num, case when f.type  = 3 then 1 else 0 end
        from discussion  d
        inner join folder f
        on d.folder_id = f.id
        where f.type  in (0, 3)
        and d.status = 0
        and d.id = $discussion_id;

    end if;

    COMMIT;

    select p.id,
    p.version,
    p.created_date,
    p.discussion_id,
    d.status discussion_status,
    p.text,
    p.user_id,
    case p.deleted when 1 then 1 else 0 end deleted,
    p.moderation_result,
    p.moderation_score,
    p.status,
    p.last_edit_date,
    case p.markdown when 1 then 1 else 0 end markdown,
    p.post_count,
    p.post_num,
    u.id user_id,
    u.username,
    case u.enabled when 1 then 1 else 0 end user_enabled,
    case u.account_locked when 1 then 1 else 0 end user_locked,
    case u.account_expired when 1 then 1 else 0 end user_expired
    from post p
    inner join discussion d
    on p.discussion_id = d.id
    inner join folder f
    on d.folder_id = f.id
    inner join user u
    on p.user_id = u.id
    where p.id = $last_post_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS edit_discussion_post;
DELIMITER //
CREATE PROCEDURE edit_discussion_post(IN $folder_id bigint, IN $discussion_id bigint, IN $post_id int, IN $body_text varchar(8192), IN $user_id int)
BEGIN

    declare $rows_affected int;
    declare $current_timestamp datetime;

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    select UTC_TIMESTAMP() into $current_timestamp;

    start transaction;

    INSERT INTO post_edit (
    version,
    created_date,
    post_id,
    text)
    select
    1,
    $current_timestamp,
    p.id,
    p.text
    from post p
    where id = $post_id;

    update post p
    inner join discussion d
    on p.discussion_id = d.id
    inner join folder f
    on d.folder_id = f.id
    set p.text = $body_text,
    p.last_edit_date = $current_timestamp
    where p.id = $post_id
    and p.user_id = $user_id
    and d.id = $discussion_id
    and f.id = $folder_id;

    SELECT ROW_COUNT() into $rows_affected;

    commit work;

    if $rows_affected = 1 then

        select p.id,
        p.version,
        p.created_date,
        p.discussion_id,
        d.status discussion_status,
        p.text,
        p.user_id,
        case p.deleted when 1 then 1 else 0 end deleted,
        p.moderation_result,
        p.moderation_score,
        p.status,
        p.last_edit_date,
        case p.markdown when 1 then 1 else 0 end markdown,
        p.post_count,
        p.post_num,
        u.id user_id,
        u.username,
        case u.enabled when 1 then 1 else 0 end user_enabled,
        case u.account_locked when 1 then 1 else 0 end user_locked,
        case u.account_expired when 1 then 1 else 0 end user_expired
        from post p
        inner join discussion d
        on p.discussion_id = d.id
        inner join folder f
        on d.folder_id = f.id
        inner join user u
        on p.user_id = u.id
        where p.id = $post_id;

    end if;


END //
DELIMITER ;

DROP PROCEDURE IF EXISTS delete_discussion_post;
DELIMITER //
CREATE PROCEDURE delete_discussion_post(IN $folder_id bigint, IN $discussion_id bigint, IN $post_id int, IN $user_id int)
BEGIN

    declare $is_admin int;
    declare $rows_affected int;
    declare $current_timestamp datetime;

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    select UTC_TIMESTAMP() into $current_timestamp;

    #static STATUS_DELETED_BY_ADMIN = 2
    #static STATUS_DELETED_BY_USER = 256

    update post p
    inner join discussion d
    on p.discussion_id = d.id
    inner join folder f
    on d.folder_id = f.id
    set p.deleted = 1,
    p.status = case when p.user_id <> $user_id then 2 else 256 end
    where p.id = $post_id
    and d.id = $discussion_id
    and f.id = $folder_id;

    SELECT ROW_COUNT() into $rows_affected;

    if $rows_affected = 1 then

        delete from moderation_queue where post_id = $post_id;

        select p.id,
        p.version,
        p.created_date,
        p.discussion_id,
        d.status discussion_status,
        p.text,
        p.user_id,
        case p.deleted when 1 then 1 else 0 end deleted,
        p.moderation_result,
        p.moderation_score,
        p.status,
        p.last_edit_date,
        case p.markdown when 1 then 1 else 0 end markdown,
        p.post_count,
        p.post_num,
        u.id user_id,
        u.username,
        case u.enabled when 1 then 1 else 0 end user_enabled,
        case u.account_locked when 1 then 1 else 0 end user_locked,
        case u.account_expired when 1 then 1 else 0 end user_expired
        from post p
        inner join discussion d
        on p.discussion_id = d.id
        inner join folder f
        on d.folder_id = f.id
        inner join user u
        on p.user_id = u.id
        where p.id = $post_id;

    end if;


END //
DELIMITER ;


DROP PROCEDURE IF EXISTS update_user_ignore;
DELIMITER //
CREATE PROCEDURE update_user_ignore(IN $user_id bigint, IN $ignored_user_id bigint, IN $state int)
BEGIN

    declare $already_exists int;
    declare $current_timestamp datetime;

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    select UTC_TIMESTAMP() into $current_timestamp;

    select count(*) into $already_exists from ignore_user where user_id = $user_id and ignored_user_id = $ignored_user_id;

    if $state = 0 then
        if $already_exists > 0 then
            delete from ignore_user where user_id = $user_id and ignored_user_id = $ignored_user_id;
        end if;
    else
        if $already_exists = 0 then
            insert into ignore_user(created_date, version, user_id, ignored_user_id)
            values ($current_timestamp, 1, $user_id, $ignored_user_id);
        end if;
    end if;

	select i.*, u.username ignored_username
    from ignore_user i
    inner join user u
    on i.ignored_user_id = u.id
    where user_id = $user_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS find_discussion_by_title;
DELIMITER //
CREATE PROCEDURE find_discussion_by_title(IN $folder_id bigint, IN $discussion_title varchar(128))
BEGIN

    select d.id,
    d.version,
    d.created_date,
    d.folder_id,
    d.header,
    d.last_post,
    d.title,
    d.user_id,
    u.username,
    case coalesce(d.deleted, 0) when 1 then 1 else 0 end deleted,
    case d.locked when 1 then 1 else 0 end locked,
    d.post_count,
    d.zorder,
    d.status,
    case d.premoderate when 1 then 1 else 0 end premoderate,
    d.last_updated,
    d.last_post_id
    from discussion d
    inner join folder f
    on d.folder_id = f.id
    inner join user u
    on d.user_id = u.id
    where d.folder_id = $folder_id
    and strcmp(d.title, $discussion_title) = 0;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS create_discussion;
DELIMITER //
CREATE PROCEDURE create_discussion(IN $folder_id bigint, IN $discussion_title varchar(128), IN $discussion_header varchar(1024), IN $user_id bigint, IN $locked int)
BEGIN

    declare $current_timestamp datetime;
    declare $discussion_id bigint;

    select UTC_TIMESTAMP() into $current_timestamp;

    start transaction;

    INSERT INTO discussion (
    version,
    created_date,
    folder_id,
    title,
    header,
    user_id,
    deleted,
    locked,
    post_count,
    zorder,
    status,
    premoderate,
    last_updated,
    last_post)
    VALUES (
    1,
    $current_timestamp,
    $folder_id,
    $discussion_title,
    $discussion_header,
    $user_id,
    0,
    $locked,
    0,
    0,
    0,
    0,
    $current_timestamp,
    $current_timestamp);

    select LAST_INSERT_ID() into $discussion_id;

    insert into front_page_entry (version, discussion_id, discussion_name, folder_id, folder_key, folder_name, last_post, last_post_id, post_count, admin_only)
    select 0, d.id, d.title, d.folder_id, f.folder_key, f.description, d.created_date, null, d.post_count, case when f.type  = 3 then 1 else 0 end
    from discussion  d
    inner join folder f
    on d.folder_id = f.id
    where d.id = $discussion_id
    and f.type  in (0, 3);

    commit work;

    call get_discussion($discussion_id);

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS edit_discussion;
DELIMITER //
CREATE PROCEDURE edit_discussion(IN $folder_id bigint, IN $discussion_id bigint, IN $discussion_title varchar(128), IN $discussion_header varchar(1024), IN $user_id bigint, IN $locked int)
BEGIN

    declare $current_timestamp datetime;

    select UTC_TIMESTAMP() into $current_timestamp;

    start transaction;

    update discussion set
    version = version + 1,
    last_updated = $current_timestamp,
    title = $discussion_title,
    header = $discussion_header,
    status = $locked
    where id = $discussion_id
    and folder_id = $folder_id;

    update front_page_entry set discussion_name = $discussion_title where discussion_id = $discussion_id;

    commit work;

    call get_discussion($discussion_id);

END //
DELIMITER ;


DROP PROCEDURE IF EXISTS delete_user_bookmark;
DELIMITER //
CREATE PROCEDURE delete_user_bookmark(IN $user_id bigint, IN $discussion_id bigint)
BEGIN

    delete from user_discussion
    where user_id = $user_id
    and discussion_id = $discussion_id;

END //
DELIMITER ;


DROP PROCEDURE IF EXISTS create_report;
DELIMITER //
CREATE PROCEDURE create_report(IN $post_id bigint, IN $user_id bigint, IN $name varchar(64), IN $email varchar(64), IN $body varchar(512), IN $ipaddress varchar(15))
BEGIN

    declare $current_timestamp datetime;

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    start transaction;

    select UTC_TIMESTAMP() into $current_timestamp;

    INSERT INTO post_report (
    version,
    comment,
    created_date,
    email,
    ipaddress,
    name,
    post_id,
    score,
    user_id)
    VALUES (
    1,
    $body,
    $current_timestamp,
    $email,
    $ipaddress,
    $name,
    $post_id,
    1,
    case $user_id when 0 then null else $user_id end);

    update post
    set moderation_score = moderation_score + 1
    where id = $post_id;

    insert into moderation_queue (
    version,
    created_date,
    post_id)
    select 1,
    $current_timestamp,
    id
    from post p
    where p.id = $post_id
    and p.id not in (select post_id from moderation_queue);

    commit work;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_moderated_posts;
DELIMITER //
CREATE PROCEDURE get_moderated_posts(IN $page_start int, IN $page_size int)
BEGIN

    select p.id,
    p.version,
    p.created_date,
    p.discussion_id,
    d.status discussion_status,
    p.text,
    p.user_id,
    case p.deleted when 1 then 1 else 0 end deleted,
    p.moderation_result,
    p.moderation_score,
    p.status,
    p.last_edit_date,
    case p.markdown when 1 then 1 else 0 end markdown,
    p.post_count,
    p.post_num,
    u.id user_id,
    u.username,
    case u.enabled when 1 then 1 else 0 end user_enabled,
    case u.account_locked when 1 then 1 else 0 end user_locked,
    case u.account_expired when 1 then 1 else 0 end user_expired,
    case coalesce(o.watch, 0) when 1 then 1 else 0 end user_watch,
    case coalesce(o.premoderate) when 1 then 1 else 0 end user_premod
    from post p
    inner join discussion d
    on p.discussion_id = d.id
    inner join folder f
    on d.folder_id = f.id
    inner join user u
    on p.user_id = u.id
    left join user_options o
    on u.id = o.user_id
    where p.moderation_score > 0
    and p.moderation_result != 0
    and p.created_date > (now() - INTERVAL 30 DAY)
    order by p.created_date desc
    limit $page_start, $page_size;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_moderation_queue;
DELIMITER //
CREATE PROCEDURE get_moderation_queue()
BEGIN

    select p.id,
    p.version,
    p.created_date,
    p.discussion_id,
    d.status discussion_status,
    p.text,
    p.user_id,
    case p.deleted when 1 then 1 else 0 end deleted,
    p.moderation_result,
    p.moderation_score,
    p.status,
    p.last_edit_date,
    case p.markdown when 1 then 1 else 0 end markdown,
    p.post_count,
    p.post_num,
    u.id user_id,
    u.username,
    case u.enabled when 1 then 1 else 0 end user_enabled,
    case u.account_locked when 1 then 1 else 0 end user_locked,
    case u.account_expired when 1 then 1 else 0 end user_expired,
    case coalesce(o.watch, 0) when 1 then 1 else 0 end user_watch,
    case coalesce(o.premoderate) when 1 then 1 else 0 end user_premod
    from post p
    inner join discussion d
    on p.discussion_id = d.id
    inner join folder f
    on d.folder_id = f.id
    inner join user u
    on p.user_id = u.id
    left join user_options o
    on u.id = o.user_id
    inner join moderation_queue mq
    on p.id = mq.post_id
    where p.status in (0, 1, 3, 4)
    order by p.created_date;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_reports_by_discussion;
DELIMITER //
CREATE PROCEDURE get_reports_by_discussion(IN $discussion_id bigint)
BEGIN

    select pr.*,
    p.status post_status,
    p.moderation_score,
    p.moderation_result,
    p.created_date post_created_date,
    p.post_num,
    p.discussion_id,
    p.user_id,
    u.username
    from post_report pr
    inner join post p
    on pr.post_id = p.id
    inner join user u
    on p.user_id = u.id
    where p.discussion_id = $discussion_id
    and p.moderation_result = 0
    order by pr.post_id, pr.created_date;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_comments_by_discussion;
DELIMITER //
CREATE PROCEDURE get_comments_by_discussion(IN $discussion_id bigint)
BEGIN


    select c.*, u.username
    from moderator_comment c
    inner join post p
    on c.post_id = p.id
    inner join user u
    on c.user_id = u.id
    where p.discussion_id = $discussion_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_reports_by_post;
DELIMITER //
CREATE PROCEDURE get_reports_by_post(IN $post_id bigint)
BEGIN

    select pr.*,
    p.status post_status,
    p.moderation_score,
    p.moderation_result,
    p.created_date post_created_date,
    p.post_num,
    p.discussion_id,
    p.user_id,
    u.username
    from post_report pr
    inner join post p
    on pr.post_id = p.id
    inner join user u
    on p.user_id = u.id
    where p.id = $post_id
    order by pr.created_date;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_comments_by_post;
DELIMITER //
CREATE PROCEDURE get_comments_by_post(IN $post_id bigint)
BEGIN


    select c.*, u.username
    from moderator_comment c
    inner join post p
    on c.post_id = p.id
    inner join user u
    on c.user_id = u.id
    where p.id = $post_id
    order by c.created_date;

END //
DELIMITER ;


DROP PROCEDURE IF EXISTS create_admin_coment;
DELIMITER //
CREATE PROCEDURE create_admin_coment(IN $post_id bigint, IN $user_id bigint, IN $body varchar(255), IN $vote int)
BEGIN


    INSERT INTO moderator_comment (
    version,
    comment,
    created_date,
    post_id,
    result,
    user_id)
    VALUES (
    1,
    $body,
    UTC_TIMESTAMP(),
    $post_id,
    $vote,
    $user_id);

    select c.*, u.username
    from moderator_comment c
    inner join user u
    on c.user_id = u.id
    where c.post_id = $post_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS lock_discussion;
DELIMITER //
CREATE PROCEDURE lock_discussion(IN $discussion_id bigint, IN $state bit)
BEGIN

    update discussion set locked = $state where id = $discussion_id;

    call get_discussion($discussion_id);

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS premoderate_discussion;
DELIMITER //
CREATE PROCEDURE premoderate_discussion(IN $discussion_id bigint, IN $state bit)
BEGIN

    update discussion set premoderate = $state where id = $discussion_id;

    call get_discussion($discussion_id);

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS delete_discussion;
DELIMITER //
CREATE PROCEDURE delete_discussion(IN $discussion_id bigint, IN $status int)
BEGIN

    start transaction;

    update discussion set status = $status, deleted = case $status when 0 then 0 else 1 end where id = $discussion_id;
    delete from front_page_entry where discussion_id = $discussion_id;

    commit work;

    call get_discussion($discussion_id);

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS move_discussion;
DELIMITER //
CREATE PROCEDURE move_discussion(IN $discussion_id bigint, IN $folder_id bigint)
BEGIN

    update discussion set folder_id = $folder_id where id = $discussion_id;

    call get_discussion($discussion_id);

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS erase_discussion;
DELIMITER //
CREATE PROCEDURE erase_discussion(IN $discussion_id bigint)
BEGIN

    delete from discussion where id = $discussion_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS check_subscriptions;
DELIMITER //
CREATE PROCEDURE check_subscriptions(IN $user_id bigint)
BEGIN

    declare $subs_fetch_order int;

    select subs_fetch_order into $subs_fetch_order from user_options where user_id = $user_id;

    if $subs_fetch_order = 0 then

        select d.id, d.last_post
        from discussion d
        inner join (select * from folder_subscription where user_id = $user_id) fs
        on d.folder_id = fs.folder_id
        left join folder_subscription_exception fse
        on d.id = fse.discussion_id
        and fs.id = fse.subscription_id
        left join (select * from user_discussion where user_id = $user_id) ud
        on d.id = ud.discussion_id
        where coalesce(d.last_post_id, 0) > coalesce(ud.last_post_id, 0)
        and datediff(UTC_TIMESTAMP(), d.last_post) < 90
        and fse.id is null
        union distinct
        select d.id, d.last_post
        from discussion d
        inner join (select * from subscription where user_id = $user_id) s
        on d.id = s.discussion_id
        left join (select * from user_discussion where user_id = $user_id) ud
        on d.id = ud.discussion_id
        where coalesce(d.last_post_id, 0) > coalesce(ud.last_post_id, 0)
        order by 2 asc;

    elseif $subs_fetch_order = 1 then

        select d.id, d.last_post
        from discussion d
        inner join (select * from folder_subscription where user_id = $user_id) fs
        on d.folder_id = fs.folder_id
        left join folder_subscription_exception fse
        on d.id = fse.discussion_id
        and fs.id = fse.subscription_id
        left join (select * from user_discussion where user_id = $user_id) ud
        on d.id = ud.discussion_id
        where coalesce(d.last_post_id, 0) > coalesce(ud.last_post_id, 0)
        and datediff(UTC_TIMESTAMP(), d.last_post) < 90
        and fse.id is null
        union distinct
        select d.id, d.last_post
        from discussion d
        inner join (select * from subscription where user_id = $user_id) s
        on d.id = s.discussion_id
        left join (select * from user_discussion where user_id = $user_id) ud
        on d.id = ud.discussion_id
        where coalesce(d.last_post_id, 0) > coalesce(ud.last_post_id, 0)
        order by 2 desc;

    elseif $subs_fetch_order = 2 then

        select d.id, d.last_post, d.post_count - coalesce(ud.last_post_count, 0) unread
        from discussion d
        inner join (select * from folder_subscription where user_id = $user_id) fs
        on d.folder_id = fs.folder_id
        left join folder_subscription_exception fse
        on d.id = fse.discussion_id
        and fs.id = fse.subscription_id
        left join (select * from user_discussion where user_id = $user_id) ud
        on d.id = ud.discussion_id
        where coalesce(d.last_post_id, 0) > coalesce(ud.last_post_id, 0)
        and datediff(UTC_TIMESTAMP(), d.last_post) < 90
        and fse.id is null
        union distinct
        select d.id, d.last_post, d.post_count - coalesce(ud.last_post_count, 0) unread
        from discussion d
        inner join (select * from subscription where user_id = $user_id) s
        on d.id = s.discussion_id
        left join (select * from user_discussion where user_id = $user_id) ud
        on d.id = ud.discussion_id
        where coalesce(d.last_post_id, 0) > coalesce(ud.last_post_id, 0)
        order by 3 desc, 2;

    end if;


END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_discussion_subscriptions;
DELIMITER //
CREATE PROCEDURE get_discussion_subscriptions(IN $user_id bigint)
BEGIN

    select s.id,
    d.id discussion_id,
    d.title discussion_name,
    f.id folder_id,
    f.folder_key,
    f.description folder_name,
    d.last_post,
    d.post_count,
    case when f.type = 0 then 0 else 1 end admin_only,
    coalesce(ud.last_post_count, 0) last_post_read_count,
    ud.last_post last_post_read_date,
    coalesce(ud.last_post_id, 0) last_post_read_id
    from subscription s
    left join user_discussion ud
    on s.discussion_id = ud.discussion_id
    and s.user_id = ud.user_id
    inner join discussion d
    on s.discussion_id = d.id
    inner join folder f
    on d.folder_id = f.id
    where s.user_id = $user_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS delete_discussion_subscription;
DELIMITER //
CREATE PROCEDURE delete_discussion_subscription(IN $user_id bigint, IN $discussion_id bigint)
BEGIN

    delete from subscription s
    where s.discussion_id = $discussion_id
    and s.user_id = $user_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_folder_subscriptions;
DELIMITER //
CREATE PROCEDURE get_folder_subscriptions(IN $user_id bigint)
BEGIN

    select *
    from folder_subscription s
    where s.user_id = $user_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS delete_folder_subscription;
DELIMITER //
CREATE PROCEDURE delete_folder_subscription(IN $user_id bigint, IN $folder_id bigint)
BEGIN

    start transaction;

    delete folder_subscription_exception
    from folder_subscription_exception
    inner join folder_subscription s
    on folder_subscription_exception.subscription_id = s.id
    where s.folder_id = $folder_id
    and s.user_id = $user_id;

    delete from folder_subscription s
    where s.folder_id = $folder_id
    and s.user_id = $user_id;

    commit work;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_folder_subscription_exceptions;
DELIMITER //
CREATE PROCEDURE get_folder_subscription_exceptions(IN $user_id bigint)
BEGIN

    select e.*
    from folder_subscription_exception e
    inner join folder_subscription s
    on e.subscription_id = s.id
    where s.user_id = $user_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_indexable_posts;
DELIMITER //
CREATE PROCEDURE get_indexable_posts()
BEGIN

    select p.id,
    p.created_date,
    p.text,
    u.username,
    d.title discussion_title,
    d.header discussion_header,
    f.description folder_name
	from post p
    inner join user u
    on p.user_id = u.id
	inner join discussion d
	on p.discussion_id = d.id
	inner join folder f
	on d.folder_id = f.id
	where p.status = 0
	and d.status = 0
	and not f.id in (33, 34);

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS create_password_reset_request;
DELIMITER //
CREATE PROCEDURE create_password_reset_request(IN $user_id bigint, IN $ip_address varchar(15), IN $reset_key varchar(128))
BEGIN

    declare $last_user_id int;

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    start transaction;

    delete from password_reset where user_id = $user_id;

    INSERT INTO password_reset (
    version,
    ip_address,
    reset_key,
    user_id,
    created_date)
    VALUES (
    1,
    $ip_address,
    $reset_key,
    $user_id,
    UTC_TIMESTAMP());

    set $last_user_id = LAST_INSERT_ID();

    commit work;

    select p.*, u.username
    from password_reset p
    inner join user u
    on p.user_id = u.id
    where p.id = $last_user_id;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS find_password_reset_request;
DELIMITER //
CREATE PROCEDURE find_password_reset_request(IN $reset_key varchar(128))
BEGIN

    select * from password_reset where reset_key = $reset_key;

END //
DELIMITER ;


DROP PROCEDURE IF EXISTS find_signup_confirmation_request;
DELIMITER //
CREATE PROCEDURE find_signup_confirmation_request(IN $confirmation_key varchar(128))
BEGIN

    select *
    from signup_confirmation
    where confirmation_key = $confirmation_key
    and ip_address is null;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS accept_signup_confirmation_request;
DELIMITER //
CREATE PROCEDURE accept_signup_confirmation_request(IN $confirmation_id int, IN $ip_address varchar(15))
BEGIN

    declare $user_id int;

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

	select user_id into $user_id from signup_confirmation where id = $confirmation_id;

    start transaction;


	update signup_confirmation
    set ip_address = $ip_address,
    last_updated = UTC_TIMESTAMP()
    where id = $confirmation_id;

    update user set account_locked = 0, email_verified = 1 where id = $user_id;

    commit work;

    call get_user($user_id);

END //
DELIMITER ;



DROP PROCEDURE IF EXISTS get_subscribers_for_post;
DELIMITER //
CREATE PROCEDURE get_subscribers_for_post(IN $post_id bigint)
BEGIN

    select ud.user_id
    from post p
    inner join discussion d
    on p.discussion_id = d.id
    inner join user_discussion ud
    on d.id = ud.discussion_id
    where p.id = $post_id
    union
    select fs.user_id
    from post p
    inner join discussion d
    on p.discussion_id = d.id
    inner join folder_subscription fs
    on d.folder_id = fs.folder_id
    left join folder_subscription_exception fse
    on d.id = fse.discussion_id
    and fs.id = fse.subscription_id
    where p.id = $post_id
    and fse.id is null;

END //
DELIMITER ;




DROP PROCEDURE IF EXISTS mark_discussion_read;
DELIMITER //
CREATE PROCEDURE mark_discussion_read(IN $user_id bigint, IN $discussion_id bigint)
BEGIN

    declare $last_post_id bigint;
    declare $last_post_count bigint;
    declare $last_post_date datetime;

    select last_post_id,
    post_count,
    last_post
    into $last_post_id,
    $last_post_count,
    $last_post_date
    from discussion
    where id = $discussion_id;

    call update_user_bookmark($user_id, $discussion_id, $last_post_id, $last_post_count, $last_post_date);

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS mark_folder_subscription_read;
DELIMITER //
CREATE PROCEDURE mark_folder_subscription_read(IN $user_id bigint, IN $subscription_id bigint)
BEGIN

    declare quit int default 0;

    declare $discussion_id bigint;
    declare $last_post_id bigint;
    declare $last_post_count bigint;
    declare $last_post_date datetime;

    declare curs
        cursor for
            select d.id,
            d.last_post_id,
            d.post_count,
            d.last_post
            from folder_subscription fs
            inner join discussion d
            on fs.folder_id = d.folder_id
            where fs.id = $subscription_id
            and fs.user_id = $user_id
            and d.last_post > fs.last_read
            and d.post_count > 0;


    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

	declare continue handler for not found set quit = 1;

    open curs;

    start transaction;

    curs_loop: loop

        fetch curs into $discussion_id, $last_post_id, $last_post_count, $last_post_date;
        if quit > 0 then
            leave curs_loop;
        end if;

        call update_user_bookmark($user_id, $discussion_id, $last_post_id, $last_post_count, $last_post_date);

    end loop curs_loop;

    update folder_subscription fs
    set fs.last_read = UTC_TIMESTAMP()
    where fs.id = $subscription_id
    and fs.user_id = $user_id;

    commit work;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS search_users;
DELIMITER //
CREATE PROCEDURE search_users(IN $search_term varchar(20))
BEGIN

    select u.id,
    u.version,
    u.email,
    u.username,
    u.created_date,
    u.last_updated,
    u.last_login_date,
    case u.account_expired when 1 then 1 else 0 end account_expired,
    case u.account_locked when 1 then 1 else 0 end account_locked,
    case u.enabled when 1 then 1 else 0 end enabled,
    case u.password_expired when 1 then 1 else 0 end password_expired,
    case o.premoderate when 1 then 1 else 0 end is_premoderate,
    case o.watch when 1 then 1 else 0 end is_watch,
    case coalesce(a.is_admin, 0) when 0 then 0 else 1 end is_admin,
    u.email_verified
    from user u
    left join user_options o
    on u.id = o.user_id
    left join (select user_id, count(*) is_admin from user_role where role_id in (2, 3) group by user_id) a
    on u.id = a.user_id
    where u.username like $search_term
    order by u.username;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS search_users_premod;
DELIMITER //
CREATE PROCEDURE search_users_premod()
BEGIN

    select u.id,
    u.version,
    u.email,
    u.username,
    u.created_date,
    u.last_updated,
    u.last_login_date,
    case u.account_expired when 1 then 1 else 0 end account_expired,
    case u.account_locked when 1 then 1 else 0 end account_locked,
    case u.enabled when 1 then 1 else 0 end enabled,
    case u.password_expired when 1 then 1 else 0 end password_expired,
    case o.premoderate when 1 then 1 else 0 end is_premoderate,
    case o.watch when 1 then 1 else 0 end is_watch,
    case coalesce(a.is_admin, 0) when 0 then 0 else 1 end is_admin,
    u.email_verified
    from user u
    left join user_options o
    on u.id = o.user_id
    left join (select user_id, count(*) is_admin from user_role where role_id in (2, 3) group by user_id) a
    on u.id = a.user_id
    where not last_login_date is null
    and o.premoderate = 1
    and u.enabled = 1
    order by u.username;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS search_users_watch;
DELIMITER //
CREATE PROCEDURE search_users_watch()
BEGIN

    select u.id,
    u.version,
    u.email,
    u.username,
    u.created_date,
    u.last_updated,
    u.last_login_date,
    case u.account_expired when 1 then 1 else 0 end account_expired,
    case u.account_locked when 1 then 1 else 0 end account_locked,
    case u.enabled when 1 then 1 else 0 end enabled,
    case u.password_expired when 1 then 1 else 0 end password_expired,
    case o.premoderate when 1 then 1 else 0 end is_premoderate,
    case o.watch when 1 then 1 else 0 end is_watch,
    case coalesce(a.is_admin, 0) when 0 then 0 else 1 end is_admin,
    u.email_verified
    from user u
    left join user_options o
    on u.id = o.user_id
    left join (select user_id, count(*) is_admin from user_role where role_id in (2, 3) group by user_id) a
    on u.id = a.user_id
    where not last_login_date is null
    and o.watch = 1
    and u.enabled = 1
    order by u.username;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS search_users_locked;
DELIMITER //
CREATE PROCEDURE search_users_locked()
BEGIN

    select u.id,
    u.version,
    u.email,
    u.username,
    u.created_date,
    u.last_updated,
    u.last_login_date,
    case u.account_expired when 1 then 1 else 0 end account_expired,
    case u.account_locked when 1 then 1 else 0 end account_locked,
    case u.enabled when 1 then 1 else 0 end enabled,
    case u.password_expired when 1 then 1 else 0 end password_expired,
    case o.premoderate when 1 then 1 else 0 end is_premoderate,
    case o.watch when 1 then 1 else 0 end is_watch,
    case coalesce(a.is_admin, 0) when 0 then 0 else 1 end is_admin,
    u.email_verified
    from user u
    left join user_options o
    on u.id = o.user_id
    left join (select user_id, count(*) is_admin from user_role where role_id in (2, 3) group by user_id) a
    on u.id = a.user_id
    where not last_login_date is null
    and u.account_locked = 1
    and u.enabled = 1
    order by u.username;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS search_users_recent;
DELIMITER //
CREATE PROCEDURE search_users_recent()
BEGIN

    select u.id,
    u.version,
    u.email,
    u.username,
    u.created_date,
    u.last_updated,
    u.last_login_date,
    case u.account_expired when 1 then 1 else 0 end account_expired,
    case u.account_locked when 1 then 1 else 0 end account_locked,
    case u.enabled when 1 then 1 else 0 end enabled,
    case u.password_expired when 1 then 1 else 0 end password_expired,
    case o.premoderate when 1 then 1 else 0 end is_premoderate,
    case o.watch when 1 then 1 else 0 end is_watch,
    case coalesce(a.is_admin, 0) when 0 then 0 else 1 end is_admin,
    u.email_verified
    from user u
    left join user_options o
    on u.id = o.user_id
    left join (select user_id, count(*) is_admin from user_role where role_id in (2, 3) group by user_id) a
    on u.id = a.user_id
    where not last_login_date is null
    and u.created_date > date_sub(now(), interval 30 day)
    order by u.username;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_user_history;
DELIMITER //
CREATE PROCEDURE get_user_history(IN $user_id bigint)
BEGIN

    select *
    from user_history
    where user_id = $user_id
    and created_date > date_sub(now(), interval 90 day)
    order by created_date desc;


END //
DELIMITER ;

DROP PROCEDURE IF EXISTS get_user_discussion_blocks;
DELIMITER //
CREATE PROCEDURE get_user_discussion_blocks()
BEGIN

    select du.id,
    d.id discussion_id,
    d.title discussion_name,
    f.id folder_id,
    f.folder_key,
    f.description folder_name,
    u.id user_id,
    u.username
    from discussion_user du
    inner join discussion d
    on du.discussion_id = d.id
    inner join folder f
    on d.folder_id = f.id
    inner join user u
    on du.user_id = u.id
    order by u.username;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS rebuild_front_page_entry;
DELIMITER //
CREATE PROCEDURE rebuild_front_page_entry()
BEGIN

    start transaction;

    delete from front_page_entry;

    insert into front_page_entry (version, discussion_id, discussion_name, folder_id, folder_key, folder_name, last_post, last_post_id, post_count, admin_only)
    select 0, d.id, d.title, d.folder_id, f.folder_key, f.description, coalesce(d.last_post, d.created_date), d.last_post_id , d.post_count, case when f.type  = 3 then 1 else 0 end
    from discussion  d
    inner join folder f
    on d.folder_id = f.id
    where f.type  in (0, 3)
    and d.status = 0
    order by last_post;

    commit work;

END //
DELIMITER ;

DROP PROCEDURE IF EXISTS fix_discussion_post_numbering;
DELIMITER //
CREATE PROCEDURE fix_discussion_post_numbering(IN $discussion_id bigint)
BEGIN

    declare quit int default 0;
    declare $post_num bigint default 0;
    declare $post_id bigint;

    declare curs
        cursor for
            select p.id
            from post p
            where discussion_id = $discussion_id
            order by p.id;


    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

	declare continue handler for not found set quit = 1;

    open curs;

    start transaction;

    curs_loop: loop

        fetch curs into $post_id;
        if quit > 0 then
            leave curs_loop;
        end if;

		set $post_num = $post_num + 1;
        update post set post_num = $post_num where id = $post_id;


    end loop curs_loop;

	update discussion set post_count = $post_num where id = $discussion_id;

    commit work;

END //
DELIMITER ;
