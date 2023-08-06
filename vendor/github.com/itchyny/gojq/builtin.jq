def not: if . then false else true end;
def in(xs): . as $x | xs | has($x);
def map(f): [.[] | f];
def with_entries(f): to_entries | map(f) | from_entries;
def select(f): if f then . else empty end;
def recurse: recurse(.[]?);
def recurse(f): def r: ., (f | r); r;
def recurse(f; cond): def r: ., (f | select(cond) | r); r;

def while(cond; update):
  def _while: if cond then ., (update | _while) else empty end;
  _while;
def until(cond; next):
  def _until: if cond then . else next | _until end;
  _until;
def repeat(f):
  def _repeat: f, _repeat;
  _repeat;
def range($end): _range(0; $end; 1);
def range($start; $end): _range($start; $end; 1);
def range($start; $end; $step): _range($start; $end; $step);

def min_by(f): _min_by(map([f]));
def max_by(f): _max_by(map([f]));
def sort_by(f): _sort_by(map([f]));
def group_by(f): _group_by(map([f]));
def unique_by(f): _unique_by(map([f]));

def arrays: select(type == "array");
def objects: select(type == "object");
def iterables: select(type | . == "array" or . == "object");
def booleans: select(type == "boolean");
def numbers: select(type == "number");
def finites: select(isfinite);
def normals: select(isnormal);
def strings: select(type == "string");
def nulls: select(. == null);
def values: select(. != null);
def scalars: select(type | . != "array" and . != "object");
def leaf_paths: paths(scalars);

def inside(xs): . as $x | xs | contains($x);
def combinations:
  if length == 0 then
    []
  else
    .[0][] as $x | [$x] + (.[1:] | combinations)
  end;
def combinations(n): [limit(n; repeat(.))] | combinations;
def walk(f):
  def _walk:
    if type == "array" then
      map(_walk)
    elif type == "object" then
      map_values(last(_walk))
    end | f;
  _walk;

def first: .[0];
def first(g): label $out | g | ., break $out;
def last: .[-1];
def last(g): reduce g as $item (null; $item);
def isempty(g): label $out | (g | false, break $out), true;
def all: all(.);
def all(y): all(.[]; y);
def all(g; y): isempty(g | select(y | not));
def any: any(.);
def any(y): any(.[]; y);
def any(g; y): isempty(g | select(y)) | not;
def limit($n; g):
  if $n > 0 then
    label $out |
    foreach g as $item (
      $n;
      . - 1;
      $item, if . <= 0 then break $out else empty end
    )
  elif $n == 0 then
    empty
  else
    g
  end;
def nth($n): .[$n];
def nth($n; g):
  if $n < 0 then
    error("nth doesn't support negative indices")
  else
    label $out |
    foreach g as $item (
      $n + 1;
      . - 1;
      if . <= 0 then $item, break $out else empty end
    )
  end;

def truncate_stream(f):
  . as $n | null | f |
  if .[0] | length > $n then .[0] |= .[$n:] else empty end;
def fromstream(f):
  { x: null, e: false } as $init |
  foreach f as $i (
    $init;
    if .e then $init end |
    if $i | length == 2 then
      setpath(["e"]; $i[0] | length == 0) |
      setpath(["x"] + $i[0]; $i[1])
    else
      setpath(["e"]; $i[0] | length == 1)
    end;
    if .e then .x else empty end
  );
def tostream:
  path(def r: (.[]? | r), .; r) as $p |
  getpath($p) |
  reduce path(.[]?) as $q ([$p, .]; [$p + $q]);

def map_values(f): .[] |= f;
def del(f): delpaths([path(f)]);
def paths: path(..) | select(. != []);
def paths(f): paths as $p | select(getpath($p) | f) | $p;

def fromdateiso8601: strptime("%Y-%m-%dT%H:%M:%S%z") | mktime;
def todateiso8601: strftime("%Y-%m-%dT%H:%M:%SZ");
def fromdate: fromdateiso8601;
def todate: todateiso8601;

def match($re): match($re; null);
def match($re; $flags): _match($re; $flags; false)[];
def test($re): test($re; null);
def test($re; $flags): _match($re; $flags; true);
def capture($re): capture($re; null);
def capture($re; $flags): match($re; $flags) | _capture;
def scan($re): scan($re; null);
def scan($re; $flags):
  match($re; $flags + "g") |
  if .captures == [] then
    .string
  else
    [.captures[].string]
  end;
def splits($re): splits($re; null);
def splits($re; $flags): split($re; $flags)[];
def sub($re; str): sub($re; str; null);
def sub($re; str; $flags):
  . as $str |
  def _sub:
    if .matches == [] then
      $str[:.offset] + .string
    else
      .matches[-1] as $r |
      {
        string: (($r | _capture | str) + $str[$r.offset+$r.length:.offset] + .string),
        offset: $r.offset,
        matches: .matches[:-1],
      } |
      _sub
    end;
  { string: "", matches: [match($re; $flags)] } | _sub;
def gsub($re; str): sub($re; str; "g");
def gsub($re; str; $flags): sub($re; str; $flags + "g");

def inputs:
  try
    repeat(input)
  catch
    if . == "break" then empty else error end;

def INDEX(stream; idx_expr):
  reduce stream as $row ({}; .[$row | idx_expr | tostring] = $row);
def INDEX(idx_expr):
  INDEX(.[]; idx_expr);
def JOIN($idx; idx_expr):
  [.[] | [., $idx[idx_expr]]];
def JOIN($idx; stream; idx_expr):
  stream | [., $idx[idx_expr]];
def JOIN($idx; stream; idx_expr; join_expr):
  stream | [., $idx[idx_expr]] | join_expr;
def IN(s): any(s == .; .);
def IN(src; s): any(src == s; .);
