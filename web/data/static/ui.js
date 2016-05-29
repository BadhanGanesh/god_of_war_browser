'use strict';

var viewPack, viewTree, viewSummary, view3d;
var dataPack, dataTree, dataSummary, data3d;

function set3dVisible(show) {
    if (show) {
        view3d.show();
        viewSummary.attr('style', '')
    } else {
        view3d.hide();
        viewSummary.attr('style', 'flex-grow:1;')
    }
}

function setTitle(viewHeh, title) {
    $(viewHeh).children(".view-item-title").text(title);
}

function packLoad() {
    dataPack.empty();
    $.getJSON('/json/pack', function(data) {
        var list = $('<ol>');
        for (var i in data.Files) {
            var fileName = data.Files[i].Name;
            list.append($('<li>')
                    .attr('filename', fileName)
                    .append($('<label>').append(fileName))
                    .append($('<a download>')
                            .addClass('button-dump')
                            .attr('href', '/dump/pack/' + fileName)) );
        }
        dataPack.append(list);
        
        $('#view-pack ol li label').click(function(ev) {
            packLoadFile($(this).parent().attr('filename'));
        });

        console.log('pack loaded');
    })
}

function packLoadFile(filename) {
    dataTree.empty();
    $.getJSON('/json/pack/' + filename, function(data) {
        var ext = filename.slice(-3).toLowerCase();
        switch (ext) {
            case 'wad':
                treeLoadWad(data);  
                break;
            default:
                dataTree.append(JSON.stringify(data, undefined, 2).replace('\n', '<br>'));
                break;
        }
        console.log('pack file ' + filename + ' loaded');
    });
}

function treeLoadWad(data) {
    var addNodes = function(nodes) {
        var ol = $('<ol>').attr('wadname', data.Name);
        for (var sn in nodes) {
            var node = data.Nodes[nodes[sn]];
            var li = $('<li>')
                    .attr('nodeid', node.Id)
                    .attr('nodeformat', node.Format)
                    .attr('nodename', node.Name)
                    .append($('<label>').append(("0000" + node.Id).substr(-4,4) + '.' + node.Name));
            
            if (node.IsLink) {
                // TODO: link visual
            } else {
                li.append($('<a download>')
                        .addClass('button-dump')
                        .attr('href', '/dump/pack/' + data.Name + '/' + node.Id))
                if (node.SubNodes) {
                    li.append(addNodes(node.SubNodes));
                }
            }
            ol.append(li);
        }
        return ol;
    }
    
    setTitle(viewTree, data.Name);
    
    if (data.Roots)
        dataTree.append(addNodes(data.Roots));
    
    $('#view-tree ol li label').click(function(ev) {
        var node_element = $(this).parent();
        
        treeLoadWadNode(dataTree.children().attr('wadname'),
                        parseInt(node_element.attr('nodeid')),
                        parseInt(node_element.attr('nodeformat')),
                        node_element.attr('nodename'));
    });
}

function treeLoadWadNode(wad, nodeid, format, nodename) {
    dataSummary.empty();
    
    $.getJSON('/json/pack/' + wad +'/' + nodeid, function(data) {
        setTitle(viewSummary, nodename);
    
        switch (format) {
            case 0x00000007: // txr
                summaryLoadWadTxr(data);
                break;
            case 0x00000008: // material
                summaryLoadWadMat(data);
                break;
            case 0x0001000f: // mesh
                summaryLoadWadMesh(data);
                break;
            case 0x0002000f: // mdl
                summaryLoadWadMdl(data);
                break;
            case 0x0000000c: // gfx pal
            default:
                set3dVisible(false);
                dataSummary.append(JSON.stringify(data, undefined, 2).replace('\n', '<br>'));
                break;
        }
        console.log('wad ' + wad + ' file ' + nodeid + ' loaded. format:' + format);
    });
}

function loadMeshFromAjax(data, textures) {
    var r_mesh = new Mesh();
    
    for (var iPart in data.Parts) {
        var part = data.Parts[iPart];
        for (var iGroup in part.Groups) {
            var group = part.Groups[iGroup]
            for (var iObject in group.Objects) {
                var object = group.Objects[iObject];
                for (var iPacket in object.Packets) {
                    var packet = object.Packets[iPacket];
                    for (var iBlock in packet.Blocks) {
                        var block = packet.Blocks[iBlock];
                        
                        var m_vertexes = [];
                        var m_indexes = [];
                        var m_colors;
                        var m_textures;
                        var m_material;

                        m_vertexes.length = block.Trias.X.length * 3;
                        
                        for (var i in block.Trias.X) {
                            var j = i * 3;
                            m_vertexes[j] = block.Trias.X[i];
                            m_vertexes[j+1] = block.Trias.Y[i];
                            m_vertexes[j+2] = block.Trias.Z[i];
                            if (!block.Trias.Skip[i]) {
                                m_indexes.push(i-1);
                                m_indexes.push(i-2);
                                m_indexes.push(i-0);
                            }
                        }
                        
                        if (block.Blend.R && block.Blend.R.length) {
                            m_colors = [];
                            m_colors.length = block.Blend.R.length * 4;
                            
                            for (var i in block.Blend.R) {
                                var j = i * 4;
                                m_colors[j] = block.Blend.R[i];
                                m_colors[j+1] = block.Blend.G[i];
                                m_colors[j+2] = block.Blend.B[i];
                                m_colors[j+3] = block.Blend.A[i];
                            }
                        }
                        
                        if (textures && object.MaterialId < textures.length) {
                            if (block.Uvs.U && block.Uvs.U.length) {
                                m_textures = [];
                                m_textures.length = block.Uvs.U.length * 2;
                                
                                m_material = textures[object.MaterialId];

                                for (var i in block.Uvs.U) {
                                    var j = i * 2;
                                    m_textures[j] = block.Uvs.U[i];
                                    m_textures[j+1] = block.Uvs.V[i];
                                }
                            }
                        }
                        
                        r_mesh.add(new MeshObject(m_vertexes, m_indexes, m_colors, m_material, m_textures));
                    }
                }
            }
        }
    }
    return r_mesh;
}

function summaryLoadWadMesh(data) {
    set3dVisible(true);
    reset3d();
        
    console.log(data, new Model(loadMeshFromAjax(data), null));
    
    redraw3d();
}

function summaryLoadWadMdl(data) {
    set3dVisible(true);
    reset3d();
    
    var table = $('<table>');
    if (data.Raw) {
        $.each(data.Raw, function(k, val) {
            switch (k) {
                case 'UnkFloats':
                case 'Someinfo':
                    val = JSON.stringify(val);
                    break;
                default:
                    break;
            }
            table.append($('<tr>').append($('<td>').append(k)));
            table.append($('<tr>').append($('<td>').append(val)));
        });
    }
    dataSummary.append(table);
    
    console.log(textureIdMap, 'before textures loading');
    
    var textrs = [];
    for (var i in data.Materials) {
        var txrs = data.Materials[i].Textures;
        if (txrs && txrs.length && txrs[0]) {
            var imgs = txrs[0].Images;
            if (imgs && imgs.length && imgs[0]) {
                textrs.push(LoadTexture(i, 'data:image/png;base64,' + imgs[0].Image));
            }
        } else {
            textrs.push(null);
        }
    }
    
    console.log(textureIdMap, 'before model loading');

    if (data.Meshes && data.Meshes.length) {
        console.log(data, new Model(loadMeshFromAjax(data.Meshes[0], textrs)));
    } else {
        console.info('no meshes in mdl', data);
    }
    
    console.log(textureIdMap, 'after loading');
    
    redraw3d();
}

function summaryLoadWadTxr(data) {
    set3dVisible(false);
    var table = $('<table>');
    $.each(data.Data, function(k, val) {
        table.append($('<tr>')
            .append($('<td>').append(k))
            .append($('<td>').append(val)));
    });
    table.append($('<tr>')
        .append($('<td>').append('Used gfx'))
        .append($('<td>').append(data.UsedGfx)));
    table.append($('<tr>')
        .append($('<td>').append('Used pal'))
        .append($('<td>').append(data.UsedPal)));

    dataSummary.append(table);
    for (var i in data.Images) {
        var img = data.Images[i];
        dataSummary.append($('<img>')
                .attr('src', 'data:image/png;base64,' + img.Image)
                .attr('alt', 'gfx:' + img.Gfx + '  pal:' + img.Pal));
    }
}

function summaryLoadWadMat(data) {
    set3dVisible(false);
    var clr = data.Mat.Color;
    var clrBgAttr = 'background-color: rgb('+parseInt(clr[0]*255)+','+parseInt(clr[1]*255)+','+parseInt(clr[2]*255)+')';

    var table = $('<table>');
    table.append($('<tr>')
        .append($('<td>').append('Color'))
        .append($('<td>').attr('style', clrBgAttr).append(
            JSON.stringify(clr, undefined, 2)
        ))
    );

    for (var l in data.Mat.Layers) {
        var layer = data.Mat.Layers[l];
        var ltable = $('<table>')

        $.each(layer, function(k, v) {
            var td = $('<td>');
            switch (k) {
                case 'Floats':
                case 'Flags':
                    td.append(JSON.stringify(v, undefined, 2));
                    break;
                case 'Texture':
                    td.append(v);
                    if (v != '') {
                        var txrobj = data.Textures[l];
                        td.append('<br>').append(txrobj.Data.GfxName);
                        td.append('<br>').append(txrobj.Data.PalName);
                        td.append('<br>').append($('<img>').attr('src', 'data:image/png;base64,' + txrobj.Images[0].Image));
                    }
                    break;
                default:
                    td.append(v);
                    break;
            }
            ltable.append($('<tr>').append($('<td>').append(k)).append(td));
        });

        table.append($('<tr>')
            .append($('<td>').append('Layer ' + (l+1)))
            .append($('<td>').append(ltable))
        );
    };

    dataSummary.append(table);
}

$(document).ready(function(){
    viewPack = $('#view-pack');
    viewTree = $('#view-tree');
    viewSummary = $('#view-summary');
    view3d = $('#view-3d');
    
    dataPack = viewPack.children('.view-item-container');
    dataTree = viewTree.children('.view-item-container');
    dataSummary = viewSummary.children('.view-item-container');
    data3d = view3d.children().children();
    
    packLoad();
    
    init3d(data3d);
});