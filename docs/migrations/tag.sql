CREATE TABLE tags (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    group_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '分组ID，0表示默认分组',
    name VARCHAR(50) NOT NULL COMMENT '标签名称',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    updated_at DATETIME NOT NULL COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    INDEX idx_group_id (group_id),
    INDEX idx_deleted_at (deleted_at),
    UNIQUE INDEX idx_group_name (group_id, name)
) COMMENT '标签表';

CREATE TABLE tag_groups (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    name VARCHAR(50) NOT NULL COMMENT '分组名称',
    sort_order INT DEFAULT 0 COMMENT '排序顺序',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    updated_at DATETIME NOT NULL COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    INDEX idx_deleted_at (deleted_at)
) COMMENT '标签分组表';