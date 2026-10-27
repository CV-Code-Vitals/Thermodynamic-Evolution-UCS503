const padding = 6;
const stubLength = 12;
const cellSize = 128;

function simplify(points) {
  const result = [];
  for (const point of points) {
    const previous = result.at(-1);
    if (previous && point[0] === previous[0] && point[1] === previous[1]) continue;
    if (result.length > 1) {
      const before = result.at(-2);
      if ((before[0] === previous[0] && previous[0] === point[0])
        || (before[1] === previous[1] && previous[1] === point[1])) result.pop();
    }
    result.push(point);
  }
  return result;
}

function ports(node, fraction) {
  const x = node.x + 16 + (node.width - 32) * fraction;
  const y = node.y + 12 + (node.height - 24) * fraction;
  return [
    [[node.x, y], [node.x - stubLength, y]],
    [[node.x + node.width, y], [node.x + node.width + stubLength, y]],
    [[x, node.y], [x, node.y - stubLength]],
    [[x, node.y + node.height], [x, node.y + node.height + stubLength]],
  ];
}

function obstacleIndex(nodes) {
  const cells = new Map();
  for (const node of nodes) {
    for (let x = Math.floor((node.x - padding) / cellSize); x <= Math.floor((node.x + node.width + padding) / cellSize); x++) {
      for (let y = Math.floor((node.y - padding) / cellSize); y <= Math.floor((node.y + node.height + padding) / cellSize); y++) {
        const key = `${x}:${y}`;
        if (!cells.has(key)) cells.set(key, []);
        cells.get(key).push(node);
      }
    }
  }
  return (points, source, target) => {
    for (let i = 1; i < points.length; i++) {
      const [a, b] = [points[i - 1], points[i]];
      const left = Math.min(a[0], b[0]), right = Math.max(a[0], b[0]);
      const top = Math.min(a[1], b[1]), bottom = Math.max(a[1], b[1]);
      for (let x = Math.floor(left / cellSize); x <= Math.floor(right / cellSize); x++) {
        for (let y = Math.floor(top / cellSize); y <= Math.floor(bottom / cellSize); y++) {
          for (const node of cells.get(`${x}:${y}`) || []) {
            const inset = node === source || node === target ? 0 : padding;
            if (right > node.x - inset && left < node.x + node.width + inset
              && bottom > node.y - inset && top < node.y + node.height + inset) return false;
          }
        }
      }
    }
    return true;
  };
}

function nearbyLanes(lanes, a, b) {
  const low = Math.min(a, b), high = Math.max(a, b), middle = (a + b) / 2;
  const distance = value => Math.max(low - value, 0, value - high);
  return [...new Set([...lanes].sort((x, y) => distance(x) - distance(y) || Math.abs(x - middle) - Math.abs(y - middle))
    .slice(0, 6).concat(lanes[0], lanes.at(-1)))];
}

export function routeGraphEdges(nodes, links) {
  if (!links.length) return [];
  const clear = obstacleIndex(nodes);
  const xLanes = [...new Set(nodes.flatMap(node => [node.x - stubLength, node.x + node.width + stubLength]))].sort((a, b) => a - b);
  const yLanes = [...new Set(nodes.flatMap(node => [node.y - stubLength, node.y + node.height + stubLength]))].sort((a, b) => a - b);
  const counts = new Map(), next = new Map(), occupied = new Map();
  for (const {source, target} of links) {
    for (const node of [source, target]) counts.set(node.id, (counts.get(node.id) || 0) + 1);
  }
  const fraction = node => {
    const index = (next.get(node.id) || 0) + 1;
    next.set(node.id, index);
    return index / (counts.get(node.id) + 1);
  };
  const segment = (a, b) => a[0] === b[0]
    ? {key:`x:${a[0]}`, low:Math.min(a[1], b[1]), high:Math.max(a[1], b[1])}
    : {key:`y:${a[1]}`, low:Math.min(a[0], b[0]), high:Math.max(a[0], b[0])};

  return links.map(link => {
    const {source, target} = link;
    const starts = ports(source, fraction(source)), ends = ports(target, fraction(target));
    let best = null, bestScore = Infinity;
    const consider = candidate => {
      const points = simplify(candidate);
      if (points.length < 2) return;
      let score = Math.max(0, points.length - 2) * 14;
      for (let i = 1; i < points.length; i++) {
        const part = segment(points[i - 1], points[i]);
        score += part.high - part.low;
        for (const previous of occupied.get(part.key) || []) {
          score += Math.max(0, Math.min(part.high, previous.high) - Math.max(part.low, previous.low)) * .8;
        }
      }
      if (score < bestScore && clear(points, source, target)) { best = points; bestScore = score; }
    };
    for (const [start, a] of starts) for (const [end, b] of ends) {
      consider([start, a, [a[0], b[1]], b, end]);
      consider([start, a, [b[0], a[1]], b, end]);
      for (const x of nearbyLanes(xLanes, a[0], b[0])) consider([start, a, [x, a[1]], [x, b[1]], b, end]);
      for (const y of nearbyLanes(yLanes, a[1], b[1])) consider([start, a, [a[0], y], [b[0], y], b, end]);
    }
    if (!best) throw new Error(`No clear connection between ${source.id} and ${target.id}`);
    for (let i = 1; i < best.length; i++) {
      const part = segment(best[i - 1], best[i]);
      if (!occupied.has(part.key)) occupied.set(part.key, []);
      occupied.get(part.key).push(part);
    }
    return {...link, points:best};
  });
}
