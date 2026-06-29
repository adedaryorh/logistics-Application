CREATE TABLE mobility_.road_nodes (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lat       DOUBLE PRECISION NOT NULL,
    lng       DOUBLE PRECISION NOT NULL,
    h3_cell   TEXT NOT NULL
);

CREATE TABLE mobility_.road_edges (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_node_id   UUID NOT NULL REFERENCES mobility_.road_nodes(id),
    to_node_id     UUID NOT NULL REFERENCES mobility_.road_nodes(id),
    distance_m     INT NOT NULL,
    weight_s       FLOAT NOT NULL,
    road_type      TEXT NOT NULL CHECK (road_type IN ('motorway','primary','secondary','residential')),
    bidirectional  BOOLEAN NOT NULL DEFAULT TRUE
);
