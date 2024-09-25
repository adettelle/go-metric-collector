create table metric
(id serial primary key , 
metric_type text not null,
metric_id varchar(30) not null, 
value double precision not null default 0,
delta bigint not null default 0, 
created_at timestamp not null default now(),
unique(metric_id, metric_type));